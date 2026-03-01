package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Resolve returns the project name using the following priority:
//  1. explicit override (non-empty override parameter)
//  2. git repository root name
//  3. current directory name
func Resolve(override string) (string, error) {
	if override != "" {
		return override, nil
	}

	// Try git root
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		root := strings.TrimSpace(string(out))
		return filepath.Base(root), nil
	}

	// Fallback: current directory name
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Base(cwd), nil
}

// VaultDir returns the path to the vault directory for the given project.
func VaultDir(project string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".envsecret", project), nil
}

// VaultPath returns the full path to vault.enc for the given project.
func VaultPath(project string) (string, error) {
	dir, err := VaultDir(project)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vault.enc"), nil
}

// BaseDir returns ~/.envsecret
func BaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".envsecret"), nil
}
