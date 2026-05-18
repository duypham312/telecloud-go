package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/hkdf"
)

const (
	MasterKeyEnv      = "TELECLOUD_MASTER_KEY"
	MasterKeyFileEnv  = "TELECLOUD_MASTER_KEY_FILE"
	EncPrefixV1       = "enc:v1:"
	KeyByteLen        = 32
	GCMNonceBytes     = 12
	defaultKeyFileName = "master.key"
)

var (
	masterKey   []byte
	masterKeyMu sync.Mutex
	// keyFileHint is set by the caller (main.go) before LoadMasterKey runs so
	// LoadMasterKey can probe a sensible default location when the env var is
	// unset. It is best-effort: an empty hint means "current working dir".
	keyFileHint     string
	keyFileHintLock sync.Mutex
)

// SetKeyFileHint records the directory where the binary would prefer to keep
// the master.key fallback file (typically alongside the SQLite database).
// Callers must invoke this before LoadMasterKey to take effect.
func SetKeyFileHint(dir string) {
	keyFileHintLock.Lock()
	defer keyFileHintLock.Unlock()
	keyFileHint = dir
}

// DefaultKeyFilePath returns the path the keystore will probe when
// TELECLOUD_MASTER_KEY is unset and TELECLOUD_MASTER_KEY_FILE is not set
// either. The hint takes precedence; empty hint falls back to ./master.key
// in the current working directory.
func DefaultKeyFilePath() string {
	if v := strings.TrimSpace(os.Getenv(MasterKeyFileEnv)); v != "" {
		return v
	}
	keyFileHintLock.Lock()
	hint := keyFileHint
	keyFileHintLock.Unlock()
	if hint != "" {
		return filepath.Join(hint, defaultKeyFileName)
	}
	return defaultKeyFileName
}

// LoadMasterKey returns the 32-byte master key. Lookup order:
// TELECLOUD_MASTER_KEY env, then a key file at DefaultKeyFilePath. The env
// value and the file accept 64-char hex or base64 (std or url-safe, padded
// or unpadded) that decodes to exactly 32 bytes.
//
// Successful loads are cached for the lifetime of the process. Errors are
// NOT cached: bootstrap may call LoadMasterKey before configuring a key,
// then again after injecting one via os.Setenv, and the second call should
// re-probe sources rather than returning the stale failure.
func LoadMasterKey() ([]byte, error) {
	masterKeyMu.Lock()
	defer masterKeyMu.Unlock()
	if masterKey != nil {
		return masterKey, nil
	}
	raw := strings.TrimSpace(os.Getenv(MasterKeyEnv))
	if raw != "" {
		key, err := DecodeKey(raw)
		if err != nil {
			return nil, fmt.Errorf("%s is invalid: %w", MasterKeyEnv, err)
		}
		masterKey = key
		return masterKey, nil
	}
	path := DefaultKeyFilePath()
	if path == "" {
		return nil, fmt.Errorf("%s is not set and no key file path is configured", MasterKeyEnv)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s is not set and no key file exists at %s", MasterKeyEnv, path)
		}
		return nil, fmt.Errorf("read master key file %s: %w", path, err)
	}
	key, err := DecodeKey(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("master key file %s is invalid: %w", path, err)
	}
	masterKey = key
	return masterKey, nil
}

// MasterKeyLoaded reports whether a master key has been successfully loaded.
func MasterKeyLoaded() bool {
	masterKeyMu.Lock()
	defer masterKeyMu.Unlock()
	return masterKey != nil
}

// DecodeKey parses a master-key string in 64-char hex or 32-byte base64
// (std/url-safe, padded or unpadded). Exposed so bootstrap can validate
// user-supplied input before persisting it.
func DecodeKey(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if b, err := hex.DecodeString(raw); err == nil && len(b) == KeyByteLen {
		return b, nil
	}
	for _, dec := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding,
		base64.URLEncoding, base64.RawURLEncoding,
	} {
		if b, err := dec.DecodeString(raw); err == nil && len(b) == KeyByteLen {
			return b, nil
		}
	}
	return nil, fmt.Errorf("expected 32 bytes encoded as hex(64 chars) or base64")
}

// GenerateMasterKey returns a fresh 32-byte key as a 64-char hex string.
func GenerateMasterKey() (string, error) {
	buf := make([]byte, KeyByteLen)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// PersistMasterKey writes the given hex-encoded master key to path with a
// 0600 mode (best-effort on Windows). The write is atomic via *.tmp + rename.
// If a key file already exists with a different value, the call fails to
// avoid silently overwriting an in-use key.
func PersistMasterKey(path, hexKey string) error {
	if path == "" {
		return errors.New("master key file path is empty")
	}
	if _, err := DecodeKey(hexKey); err != nil {
		return fmt.Errorf("refuse to persist invalid key: %w", err)
	}
	if existing, err := os.ReadFile(path); err == nil {
		existingStr := strings.TrimSpace(string(existing))
		if existingStr != "" && existingStr != strings.TrimSpace(hexKey) {
			return fmt.Errorf("master key file already exists at %s with a different value; refuse to overwrite", path)
		}
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("create key dir %s: %w", dir, err)
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.TrimSpace(hexKey)+"\n"), 0600); err != nil {
		return fmt.Errorf("write key tmp: %w", err)
	}
	// Best-effort chmod (Windows ignores).
	_ = os.Chmod(tmp, 0600)
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename key file: %w", err)
	}
	return nil
}

// DeriveSubKey returns a 32-byte sub-key derived from the master key using HKDF-SHA256
// with the given label as the info parameter. Useful for separating concerns (HMAC vs AEAD vs ...).
func DeriveSubKey(label string) ([]byte, error) {
	mk, err := LoadMasterKey()
	if err != nil {
		return nil, err
	}
	r := hkdf.New(sha256.New, mk, nil, []byte(label))
	out := make([]byte, 32)
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, err
	}
	return out, nil
}

// EncryptAEAD encrypts plaintext with AES-256-GCM using the master key.
// Output layout: nonce(12) || ciphertext || tag(16).
func EncryptAEAD(plaintext []byte) ([]byte, error) {
	mk, err := LoadMasterKey()
	if err != nil {
		return nil, err
	}
	return EncryptAEADWith(mk, plaintext)
}

// EncryptAEADWith encrypts plaintext with AES-256-GCM using the provided
// 32-byte key. Bootstrap uses this to verify a freshly-supplied key without
// committing it to the package-level cache.
func EncryptAEADWith(key, plaintext []byte) ([]byte, error) {
	if len(key) != KeyByteLen {
		return nil, fmt.Errorf("key must be %d bytes", KeyByteLen)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, len(nonce)+len(ct))
	out = append(out, nonce...)
	out = append(out, ct...)
	return out, nil
}

// DecryptAEAD reverses EncryptAEAD.
func DecryptAEAD(blob []byte) ([]byte, error) {
	mk, err := LoadMasterKey()
	if err != nil {
		return nil, err
	}
	return DecryptAEADWith(mk, blob)
}

// DecryptAEADWith reverses EncryptAEAD with the supplied key. Bootstrap uses
// it to validate a pasted key against an existing ciphertext row before
// committing the key to disk.
func DecryptAEADWith(key, blob []byte) ([]byte, error) {
	if len(key) != KeyByteLen {
		return nil, fmt.Errorf("key must be %d bytes", KeyByteLen)
	}
	if len(blob) < GCMNonceBytes+16 {
		return nil, errors.New("ciphertext too short")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := blob[:gcm.NonceSize()]
	ct := blob[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}

// EncryptString returns "enc:v1:<base64(EncryptAEAD(s))>".
func EncryptString(s string) (string, error) {
	enc, err := EncryptAEAD([]byte(s))
	if err != nil {
		return "", err
	}
	return EncPrefixV1 + base64.RawStdEncoding.EncodeToString(enc), nil
}

// DecryptString reverses EncryptString. Values without the "enc:v1:" prefix are
// returned as-is (so callers can read legacy plaintext rows; auto-migration is
// responsible for re-encrypting them).
func DecryptString(s string) (string, error) {
	if !strings.HasPrefix(s, EncPrefixV1) {
		return s, nil
	}
	raw, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(s, EncPrefixV1))
	if err != nil {
		return "", err
	}
	plain, err := DecryptAEAD(raw)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// IsEncryptedString reports whether s carries the encryption prefix.
func IsEncryptedString(s string) bool {
	return strings.HasPrefix(s, EncPrefixV1)
}

// DecodeEncryptedString decodes the "enc:v1:" envelope without invoking the
// cached master key. Bootstrap uses this together with DecryptAEADWith to
// validate a candidate key against an existing ciphertext row.
func DecodeEncryptedString(s string) ([]byte, error) {
	if !strings.HasPrefix(s, EncPrefixV1) {
		return nil, errors.New("missing enc:v1: prefix")
	}
	return base64.RawStdEncoding.DecodeString(strings.TrimPrefix(s, EncPrefixV1))
}
