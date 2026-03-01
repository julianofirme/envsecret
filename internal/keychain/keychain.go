// Package keychain provides a thin wrapper around the OS keychain.
// On macOS it uses the macOS Keychain; on Linux it uses libsecret via
// the go-keyring library.
package keychain

import (
	"encoding/hex"
	"fmt"

	gokeyring "github.com/zalando/go-keyring"
)

const service = "envsecret"

// keychainAccount returns the keychain account name for a project.
func keychainAccount(project string) string {
	return "project:" + project
}

// Get retrieves the master key for the given project from the OS keychain.
// Returns the raw key bytes.
func Get(project string) ([]byte, error) {
	hexKey, err := gokeyring.Get(service, keychainAccount(project))
	if err != nil {
		return nil, fmt.Errorf("keychain get [%s]: %w", project, err)
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("keychain: corrupt key for project %s", project)
	}
	return key, nil
}

// Set stores masterKey in the OS keychain for the given project.
func Set(project string, masterKey []byte) error {
	hexKey := hex.EncodeToString(masterKey)
	if err := gokeyring.Set(service, keychainAccount(project), hexKey); err != nil {
		return fmt.Errorf("keychain set [%s]: %w", project, err)
	}
	return nil
}

// Delete removes the master key for the given project from the OS keychain.
func Delete(project string) error {
	return gokeyring.Delete(service, keychainAccount(project))
}

// Exists returns true if a key exists in the keychain for this project.
func Exists(project string) bool {
	_, err := gokeyring.Get(service, keychainAccount(project))
	return err == nil
}
