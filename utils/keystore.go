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
	"strings"
	"sync"

	"golang.org/x/crypto/hkdf"
)

const (
	masterKeyEnv  = "TELECLOUD_MASTER_KEY"
	encPrefixV1   = "enc:v1:"
	keyByteLen    = 32
	gcmNonceBytes = 12
)

var (
	masterKey     []byte
	masterKeyOnce sync.Once
	masterKeyErr  error
)

// LoadMasterKey reads the TELECLOUD_MASTER_KEY env var and caches the decoded bytes.
// The env value is accepted as 64-char hex, or base64 (std or url-safe, with or without padding),
// and must decode to exactly 32 bytes.
func LoadMasterKey() ([]byte, error) {
	masterKeyOnce.Do(func() {
		raw := strings.TrimSpace(os.Getenv(masterKeyEnv))
		if raw == "" {
			masterKeyErr = fmt.Errorf("%s is not set. Generate one with: openssl rand -hex 32", masterKeyEnv)
			return
		}
		key, err := decodeKey(raw)
		if err != nil {
			masterKeyErr = fmt.Errorf("%s is invalid: %w", masterKeyEnv, err)
			return
		}
		masterKey = key
	})
	if masterKeyErr != nil {
		return nil, masterKeyErr
	}
	return masterKey, nil
}

// MasterKeyLoaded reports whether a master key has been successfully loaded.
func MasterKeyLoaded() bool {
	return masterKey != nil
}

func decodeKey(raw string) ([]byte, error) {
	if b, err := hex.DecodeString(raw); err == nil && len(b) == keyByteLen {
		return b, nil
	}
	for _, dec := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding,
		base64.URLEncoding, base64.RawURLEncoding,
	} {
		if b, err := dec.DecodeString(raw); err == nil && len(b) == keyByteLen {
			return b, nil
		}
	}
	return nil, fmt.Errorf("expected 32 bytes encoded as hex(64 chars) or base64")
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
	block, err := aes.NewCipher(mk)
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
	if len(blob) < gcmNonceBytes+16 {
		return nil, errors.New("ciphertext too short")
	}
	mk, err := LoadMasterKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(mk)
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
	return encPrefixV1 + base64.RawStdEncoding.EncodeToString(enc), nil
}

// DecryptString reverses EncryptString. Values without the "enc:v1:" prefix are
// returned as-is (so callers can read legacy plaintext rows; auto-migration is
// responsible for re-encrypting them).
func DecryptString(s string) (string, error) {
	if !strings.HasPrefix(s, encPrefixV1) {
		return s, nil
	}
	raw, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(s, encPrefixV1))
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
	return strings.HasPrefix(s, encPrefixV1)
}
