// Package project_test provides black-box tests for the project package.
package project_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/julianofirme/envsecret/internal/project"
)

// TestResolveExplicitOverride verifies that a non-empty override is returned
// as-is, regardless of the working directory or git state.
func TestResolveExplicitOverride(t *testing.T) {
	got, err := project.Resolve("my-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my-app" {
		t.Errorf("got %q, want %q", got, "my-app")
	}
}

// TestResolveCwdFallback verifies that when no override is given and the
// working directory is not inside a git repository, the current directory
// name is returned.
func TestResolveCwdFallback(t *testing.T) {
	// Create a temp dir that is NOT a git repo.
	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	got, err := project.Resolve("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The result should be the base name of dir (os.TempDir() returns an
	// absolute path; filepath.Base gives us the last component).
	want := filepath.Base(dir)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestResolveGitRoot verifies that Resolve returns the git repository root's
// base name when called from inside a git repo.
func TestResolveGitRoot(t *testing.T) {
	// Create a temporary directory and initialise a git repo inside it.
	dir := t.TempDir()
	repoName := "myrepo"
	repoDir := filepath.Join(dir, repoName)
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	// git init
	if err := runGit(t, repoDir, "init"); err != nil {
		t.Skipf("git not available: %v", err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	got, err := project.Resolve("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != repoName {
		t.Errorf("got %q, want %q", got, repoName)
	}
}

// TestVaultDir verifies VaultDir returns ~/.envsecret/<project>.
func TestVaultDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	want := filepath.Join(home, ".envsecret", "testproject")

	got, err := project.VaultDir("testproject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestVaultPath verifies VaultPath returns ~/.envsecret/<project>/vault.enc.
func TestVaultPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	want := filepath.Join(home, ".envsecret", "testproject", "vault.enc")

	got, err := project.VaultPath("testproject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestBaseDir verifies BaseDir returns ~/.envsecret.
func TestBaseDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	want := filepath.Join(home, ".envsecret")

	got, err := project.BaseDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestVaultPathContainsProject verifies that VaultPath embeds the project name.
func TestVaultPathContainsProject(t *testing.T) {
	proj := "alpha-bravo"
	got, err := project.VaultPath(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, proj) {
		t.Errorf("VaultPath %q does not contain project name %q", got, proj)
	}
	if filepath.Base(got) != "vault.enc" {
		t.Errorf("expected filename vault.enc, got %q", filepath.Base(got))
	}
}

// runGit is a small helper that executes a git sub-command in dir.
func runGit(t *testing.T, dir string, args ...string) error {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}
