// Package bootstrap provides the first-run UX for configuring the master
// encryption key when TELECLOUD_MASTER_KEY is unset and no key file exists.
//
// When the operator is upgrading an older install or just running the binary
// for the first time, fatal-erroring on the missing env var produces an
// unrecoverable boot loop. Instead, bootstrap stands up a minimal,
// loopback-only HTTP server that walks the admin through generating or
// pasting a key, validates it against any existing encrypted data, and
// persists it to a key file before handing control back to the main startup
// sequence.
//
// Strict-mode operators (TELECLOUD_REQUIRE_MASTER_KEY=1) keep the original
// fail-fast behavior.
package bootstrap

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"telecloud/config"
	"telecloud/database"
	"telecloud/utils"
)

// EnsureMasterKey makes sure a usable master key is available before the rest
// of startup runs. Lookup order:
//
//  1. TELECLOUD_MASTER_KEY env var (already handled by utils.LoadMasterKey).
//  2. Key file at cfg.MasterKeyFile or utils.DefaultKeyFilePath.
//  3. If cfg.RequireMasterKey is set, return an error (strict mode).
//  4. Otherwise run an interactive bootstrap web UI on cfg.ListenAddr:cfg.Port
//     and block until the admin supplies a key.
//
// On success, the master key is loaded into utils' cache and the rest of the
// app can call utils.EncryptAEAD/DecryptAEAD as usual.
func EnsureMasterKey(cfg *config.Config, webFS fs.FS) error {
	// Tell the keystore where to probe for the fallback file. cfg.MasterKeyFile
	// (TELECLOUD_MASTER_KEY_FILE) wins; otherwise we hint at the database dir
	// so the key lives next to whatever the operator already backs up.
	utils.SetKeyFileHint(defaultKeyDir(cfg))

	if _, err := utils.LoadMasterKey(); err == nil {
		log.Printf("Master key loaded (source: %s).", describeKeySource(cfg))
		return nil
	}

	// No env, no file. Strict deployments want this to fail loudly.
	if cfg.RequireMasterKey {
		return fmt.Errorf("TELECLOUD_MASTER_KEY is not configured and TELECLOUD_REQUIRE_MASTER_KEY=1 forbids auto-bootstrap. Set the env var (or %s) and restart", utils.MasterKeyFileEnv)
	}

	keyPath := resolveKeyPath(cfg)
	log.Println("================================================================")
	log.Println("[bootstrap] TELECLOUD_MASTER_KEY is not configured.")
	log.Println("[bootstrap] Starting one-time setup server to configure the key.")
	log.Printf("[bootstrap] Key will be persisted to: %s", keyPath)
	log.Println("[bootstrap] Tip: set TELECLOUD_REQUIRE_MASTER_KEY=1 in production to disable this fallback.")
	log.Println("================================================================")

	hexKey, err := runBootstrapServer(cfg, webFS, keyPath)
	if err != nil {
		return fmt.Errorf("bootstrap server: %w", err)
	}

	if err := utils.PersistMasterKey(keyPath, hexKey); err != nil {
		return fmt.Errorf("persist master key: %w", err)
	}

	// Inject into the env so the cached LoadMasterKey lookup (about to run for
	// the first time) finds it without needing to peek at the file we just
	// wrote.
	_ = os.Setenv(utils.MasterKeyEnv, hexKey)

	if _, err := utils.LoadMasterKey(); err != nil {
		return fmt.Errorf("loaded fresh key but keystore rejected it: %w", err)
	}
	log.Printf("[bootstrap] Master key saved to %s and loaded into the keystore.", keyPath)
	return nil
}

func resolveKeyPath(cfg *config.Config) string {
	if cfg.MasterKeyFile != "" {
		return cfg.MasterKeyFile
	}
	return filepath.Join(defaultKeyDir(cfg), "master.key")
}

func defaultKeyDir(cfg *config.Config) string {
	if cfg.MasterKeyFile != "" {
		dir := filepath.Dir(cfg.MasterKeyFile)
		if dir == "" {
			return "."
		}
		return dir
	}
	driver := strings.ToLower(strings.TrimSpace(cfg.DatabaseDriver))
	if driver == "" || driver == "sqlite" {
		dir := filepath.Dir(cfg.DatabasePath)
		if dir == "" {
			return "."
		}
		return dir
	}
	return "."
}

func describeKeySource(cfg *config.Config) string {
	if strings.TrimSpace(os.Getenv(utils.MasterKeyEnv)) != "" {
		return "env"
	}
	return "file " + resolveKeyPath(cfg)
}

// hasExistingEncryptedData reports whether the database already holds rows
// encrypted under some master key. When true, the bootstrap UI must refuse
// "Generate new key" and force the admin to paste the original key —
// otherwise the new key would orphan the existing ciphertext.
//
// The check is implemented by probing for any actual ciphertext rather than
// trusting a schema_version flag: an install can have schema_version=1 with
// zero ciphertext (e.g. it was provisioned but never used), in which case
// generating a fresh key is still safe.
func hasExistingEncryptedData() (bool, error) {
	probe, err := findEncryptedProbe()
	if err != nil {
		return false, err
	}
	return probe != nil, nil
}

// findEncryptedProbe returns one ciphertext blob already stored in the DB so
// a paste-flow caller can confirm the supplied key actually decrypts existing
// data. The caller treats (nil, nil) as "nothing to validate against".
//
// Detection order:
//
//  1. Sensitive settings: an explicit "enc:v1:" prefix definitively marks
//     ciphertext.
//  2. tg_sessions: blobs have no prefix, so the value alone is ambiguous
//     (a pre-encryption upgrade still has plaintext session data sitting in
//     the same column). We only treat a non-empty blob as a probe when the
//     encryption schema_version has already been bumped — i.e. a previous
//     run of MigrateEncryptV1 actually wrote ciphertext.
func findEncryptedProbe() ([]byte, error) {
	for _, k := range database.SensitiveSettingKeys() {
		raw := database.GetSettingRaw(k)
		if !utils.IsEncryptedString(raw) {
			continue
		}
		blob, err := utils.DecodeEncryptedString(raw)
		if err != nil {
			continue
		}
		return blob, nil
	}
	if database.RODB == nil {
		return nil, nil
	}
	var v int
	if err := database.RODB.Get(&v, "SELECT version FROM schema_version WHERE id = ?", "encryption"); err != nil {
		// Table missing or row absent → encryption migration never ran, so
		// any tg_sessions data is plaintext. Nothing to probe against.
		return nil, nil
	}
	if v < 1 {
		return nil, nil
	}
	var data []byte
	if err := database.RODB.Get(&data, "SELECT data FROM tg_sessions LIMIT 1"); err != nil || len(data) == 0 {
		return nil, nil
	}
	return data, nil
}

func validateAgainstExisting(key []byte) error {
	probe, err := findEncryptedProbe()
	if err != nil {
		return err
	}
	if probe == nil {
		return nil
	}
	if _, err := utils.DecryptAEADWith(key, probe); err != nil {
		return errors.New("key does not match the existing encrypted data in the database")
	}
	return nil
}

func generateBootstrapToken() string {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		// crypto/rand only fails on catastrophic OS errors; fall back to a
		// timestamp-derived value so we still produce a token.
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(buf)
}

const bootstrapCookieName = "bootstrap_token_ok"

type bootstrapState struct {
	mu      sync.Mutex
	saved   bool
	hexKey  string
	keyPath string
}

func runBootstrapServer(cfg *config.Config, webFS fs.FS, keyPath string) (string, error) {
	tpl, err := template.ParseFS(webFS, "templates/bootstrap.html")
	if err != nil {
		return "", fmt.Errorf("parse bootstrap template: %w", err)
	}

	// Serve the same embedded static assets the main router uses so the
	// bootstrap template can pull in tailwind.css, style.min.css, the i18n
	// loader, fonts and locale JSON. Without this every <link> would 404
	// and the page would render without the design system.
	staticFS, staticErr := fs.Sub(webFS, "static")

	token := generateBootstrapToken()
	listenAddr := cfg.ListenAddr
	if listenAddr == "" {
		listenAddr = "127.0.0.1"
	}
	addr := listenAddr + ":" + cfg.Port

	bootURL := fmt.Sprintf("http://127.0.0.1:%s/bootstrap?token=%s", cfg.Port, token)
	log.Println("================================================================")
	log.Println("[bootstrap] Open this URL in a browser to finish setup:")
	log.Printf("[bootstrap]     %s", bootURL)
	if listenAddr != "127.0.0.1" {
		log.Printf("[bootstrap] (listening on %s — make sure that interface is private)", listenAddr)
	}
	log.Println("================================================================")

	state := &bootstrapState{keyPath: keyPath}
	done := make(chan struct{})
	var doneOnce sync.Once
	notifyDone := func() { doneOnce.Do(func() { close(done) }) }

	mux := http.NewServeMux()
	if staticErr == nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}
	mux.HandleFunc("/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		if !verifyBootstrapToken(w, r, token) {
			return
		}
		hasData, _ := hasExistingEncryptedData()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.Execute(w, map[string]interface{}{
			"KeyPath":           keyPath,
			"HasEncryptedData":  hasData,
			"DisplayListenAddr": listenAddr,
			"Port":              cfg.Port,
			"Version":           cfg.Version,
		})
	})

	mux.HandleFunc("/bootstrap/status", func(w http.ResponseWriter, r *http.Request) {
		if !verifyBootstrapToken(w, r, token) {
			return
		}
		hasData, _ := hasExistingEncryptedData()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"has_encrypted_data": hasData,
			"key_path":           keyPath,
		})
	})

	mux.HandleFunc("/bootstrap/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !verifyBootstrapToken(w, r, token) {
			return
		}
		hasData, err := hasExistingEncryptedData()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		if hasData {
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": "refusing to generate a new key: the database already contains encrypted data. Paste the original key instead.",
			})
			return
		}
		hexKey, err := utils.GenerateMasterKey()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"key": hexKey})
	})

	mux.HandleFunc("/bootstrap/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !verifyBootstrapToken(w, r, token) {
			return
		}
		var req struct {
			Key              string `json:"key"`
			ConfirmBackedUp  bool   `json:"confirm_backed_up"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		if !req.ConfirmBackedUp {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "confirm_backed_up is required"})
			return
		}
		decoded, err := utils.DecodeKey(strings.TrimSpace(req.Key))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if err := validateAgainstExisting(decoded); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		state.mu.Lock()
		if state.saved {
			state.mu.Unlock()
			writeJSON(w, http.StatusConflict, map[string]string{"error": "already saved"})
			return
		}
		state.saved = true
		state.hexKey = strings.TrimSpace(req.Key)
		state.mu.Unlock()

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":       true,
			"key_path": keyPath,
		})

		go func() {
			time.Sleep(300 * time.Millisecond)
			notifyDone()
		}()
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Anything else: nudge the caller toward /bootstrap.
		w.Header().Set("Retry-After", "5")
		http.Error(w, "TeleCloud is waiting for master-key bootstrap. Visit /bootstrap to continue.", http.StatusServiceUnavailable)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return "", err
		}
		return "", errors.New("bootstrap server exited before a key was supplied")
	case <-done:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)

	state.mu.Lock()
	hexKey := state.hexKey
	state.mu.Unlock()
	if hexKey == "" {
		return "", errors.New("internal: bootstrap completed without a key")
	}
	return hexKey, nil
}

func verifyBootstrapToken(w http.ResponseWriter, r *http.Request, expected string) bool {
	if cookie, err := r.Cookie(bootstrapCookieName); err == nil && constantEq(cookie.Value, expected) {
		return true
	}
	supplied := r.Header.Get("X-Bootstrap-Token")
	if supplied == "" {
		supplied = r.URL.Query().Get("token")
	}
	if supplied != "" && constantEq(supplied, expected) {
		http.SetCookie(w, &http.Cookie{
			Name:     bootstrapCookieName,
			Value:    expected,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   3600,
		})
		return true
	}
	http.Error(w, "missing or invalid bootstrap token (see server logs for the URL)", http.StatusForbidden)
	return false
}

func constantEq(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
