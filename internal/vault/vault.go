// Package vault handles AES-256-GCM encryption and decryption of the secret
// store.  The vault file format is:
//
//	[4 bytes: version=1] [32 bytes: scrypt salt] [12 bytes: GCM nonce] [ciphertext+tag]
//
// The plaintext is a newline-delimited list of KEY=VALUE pairs (JSON encoding
// of a map[string]string is used so values can contain arbitrary bytes).
package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/scrypt"
)

const (
	version   uint32 = 1
	saltLen          = 32
	nonceLen         = 12
	headerLen        = 4 + saltLen + nonceLen // version + salt + nonce
	scryptN          = 16384
	scryptR          = 8
	scryptP          = 1
	keyLen           = 32
)

// deriveKey derives a 256-bit AES key from masterKey using scrypt + salt.
func deriveKey(masterKey []byte, salt []byte) ([]byte, error) {
	return scrypt.Key(masterKey, salt, scryptN, scryptR, scryptP, keyLen)
}

// Load reads and decrypts the vault at path, returning the secrets map.
// Returns an empty map if the file does not exist.
func Load(path string, masterKey []byte) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read vault: %w", err)
	}
	return decrypt(data, masterKey)
}

// Save encrypts secrets and writes them to path with mode 0600.
func Save(path string, masterKey []byte, secrets map[string]string) error {
	data, err := encrypt(masterKey, secrets)
	if err != nil {
		return err
	}
	// Write atomically via temp file
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write vault: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename vault: %w", err)
	}
	// Ensure permissions even if pre-existing file had looser perms
	return os.Chmod(path, 0600)
}

func encrypt(masterKey []byte, secrets map[string]string) ([]byte, error) {
	plaintext, err := json.Marshal(secrets)
	if err != nil {
		return nil, fmt.Errorf("marshal secrets: %w", err)
	}

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("rand salt: %w", err)
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("rand nonce: %w", err)
	}

	dk, err := deriveKey(masterKey, salt)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}

	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	out := make([]byte, headerLen+len(ciphertext))
	binary.BigEndian.PutUint32(out[0:4], version)
	copy(out[4:4+saltLen], salt)
	copy(out[4+saltLen:headerLen], nonce)
	copy(out[headerLen:], ciphertext)
	return out, nil
}

func decrypt(data []byte, masterKey []byte) (map[string]string, error) {
	if len(data) < headerLen+16 { // 16 = min GCM tag
		return nil, errors.New("vault: file too short")
	}
	v := binary.BigEndian.Uint32(data[0:4])
	if v != version {
		return nil, fmt.Errorf("vault: unknown version %d", v)
	}
	salt := data[4 : 4+saltLen]
	nonce := data[4+saltLen : headerLen]
	ciphertext := data[headerLen:]

	dk, err := deriveKey(masterKey, salt)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}

	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("vault: decryption failed (wrong key or tampered data)")
	}

	var secrets map[string]string
	if err := json.Unmarshal(plaintext, &secrets); err != nil {
		return nil, fmt.Errorf("vault: unmarshal: %w", err)
	}
	return secrets, nil
}
