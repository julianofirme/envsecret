package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/project"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all projects with initialized vaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, err := project.BaseDir()
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(baseDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("0 project(s).")
				return nil
			}
			return fmt.Errorf("read base dir: %w", err)
		}

		var projects []string
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			vaultFile := filepath.Join(baseDir, e.Name(), "vault.enc")
			if _, err := os.Stat(vaultFile); err == nil {
				projects = append(projects, e.Name())
			}
		}

		sort.Strings(projects)

		// Detect current project
		currentProj := ""
		out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err == nil {
			currentProj = filepath.Base(strings.TrimSpace(string(out)))
		} else {
			cwd, _ := os.Getwd()
			currentProj = filepath.Base(cwd)
		}

		fmt.Printf("%d project(s):\n", len(projects))
		if len(projects) > 0 {
			fmt.Println()
			for _, p := range projects {
				if p == currentProj {
					fmt.Printf("  %s (current)\n", p)
				} else {
					fmt.Printf("  %s\n", p)
				}
			}
		}
		return nil
	},
}
