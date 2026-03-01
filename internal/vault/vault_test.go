// Package vault_test provides black-box tests for the vault package.
package vault_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/julianofirme/envsecret/internal/vault"
)

var testKey = []byte("01234567890123456789012345678901") // 32 bytes

// roundTrip is a helper that saves secrets and loads them back.
func roundTrip(t *testing.T, secrets map[string]string) map[string]string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := vault.Save(path, testKey, secrets); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := vault.Load(path, testKey)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return got
}

// TestRoundTrip verifies that a non-empty map survives a save/load cycle.
func TestRoundTrip(t *testing.T) {
	want := map[string]string{
		"FOO":     "bar",
		"DB_HOST": "localhost",
		"SPECIAL": "hello=world\nnewline",
	}
	got := roundTrip(t, want)

	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %d, want %d", len(got), len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("key %q: got %q, want %q", k, got[k], v)
		}
	}
}

// TestEmptyMap verifies that an empty map round-trips cleanly.
func TestEmptyMap(t *testing.T) {
	got := roundTrip(t, map[string]string{})
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %v", got)
	}
}

// TestLoadMissingFile verifies that Load returns an empty map (not an error)
// when the vault file does not exist.
func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.enc")

	got, err := vault.Load(path, testKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map for missing file, got %v", got)
	}
}

// TestWrongKey verifies that decryption with the wrong master key returns an error.
func TestWrongKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := vault.Save(path, testKey, map[string]string{"K": "V"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	wrongKey := []byte("99999999999999999999999999999999")
	_, err := vault.Load(path, wrongKey)
	if err == nil {
		t.Fatal("expected error when loading with wrong key, got nil")
	}
}

// TestTamperedData verifies that a bit-flip in the ciphertext causes Load to fail.
func TestTamperedData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := vault.Save(path, testKey, map[string]string{"K": "V"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	// Flip a byte in the ciphertext area (after the 48-byte header).
	data[len(data)-1] ^= 0xFF
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err = vault.Load(path, testKey)
	if err == nil {
		t.Fatal("expected error for tampered data, got nil")
	}
}

// TestTruncatedFile verifies that a file shorter than the minimum header
// returns an error instead of panicking.
func TestTruncatedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	// Write only 10 bytes — far shorter than the 48-byte header + 16-byte GCM tag.
	if err := os.WriteFile(path, make([]byte, 10), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := vault.Load(path, testKey)
	if err == nil {
		t.Fatal("expected error for truncated file, got nil")
	}
}

// TestUnknownVersion verifies that a vault with an unrecognised version field
// returns an error.
func TestUnknownVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := vault.Save(path, testKey, map[string]string{"K": "V"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	// Overwrite the 4-byte version field with 99.
	binary.BigEndian.PutUint32(data[0:4], 99)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err = vault.Load(path, testKey)
	if err == nil {
		t.Fatal("expected error for unknown version, got nil")
	}
}

// TestFilePermissions verifies that Save always writes the vault with mode 0600.
func TestFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := vault.Save(path, testKey, map[string]string{"K": "V"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("permissions: got %04o, want 0600", perm)
	}
}

// TestAtomicWrite verifies that a .tmp file is not left behind after a
// successful Save.
func TestAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	if err := vault.Save(path, testKey, map[string]string{"K": "V"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	tmp := path + ".tmp"
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Errorf("expected .tmp file to be absent after Save, but Stat returned: %v", err)
	}
}

// TestFreshSaltAndNonce verifies that two consecutive saves of the same data
// produce different ciphertext (because salt and nonce are re-randomised).
func TestFreshSaltAndNonce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")
	secrets := map[string]string{"K": "V"}

	if err := vault.Save(path, testKey, secrets); err != nil {
		t.Fatalf("first Save: %v", err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile 1: %v", err)
	}

	if err := vault.Save(path, testKey, secrets); err != nil {
		t.Fatalf("second Save: %v", err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile 2: %v", err)
	}

	if string(first) == string(second) {
		t.Error("two consecutive saves produced identical ciphertext — salt/nonce not re-randomised")
	}
}
