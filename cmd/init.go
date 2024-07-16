package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	esRootDirName       = ".es"
	statusFilePath      = filepath.Join(esRootDirName, "status.txt")
	stagingAreaFilePath = filepath.Join(esRootDirName, "staging-area.txt")
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "This command initializes the environment variable tracking system",
	Example: "es init",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runInitCommand()
	},
}

func runInitCommand() error {
	if exists, err := checkPathExists(esRootDirName); err != nil {
		return err
	} else if exists {
		return errors.New("es root directory already exists")
	}

	if err := createDirectory(esRootDirName); err != nil {
		return err
	}

	if err := createFile(statusFilePath); err != nil {
		return err
	}

	if err := createFile(stagingAreaFilePath); err != nil {
		return err
	}

	color.Green("Environment variable tracking system initialized in .es/ directory!")

	return nil
}

func checkPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func createDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func createFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}
