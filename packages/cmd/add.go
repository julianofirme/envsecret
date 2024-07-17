package cmd

import (
	"bufio"
	"envsecret/packages/util"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type envVarMetadata struct {
	Key          string
	Value        string
	Modification string
}

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "This command will track changes in your .env file",
	PreRun: func(cmd *cobra.Command, args []string) {
		util.RequireLogin()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAddCommand(stagingAreaFilePath, args)
	},
}

func runAddCommand(stagingAreaFilePath string, filePaths []string) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("no .env file path provided")
	}

	envFilePath := filePaths[0]
	previousEnvVars, err := loadEnvVars(statusFilePath)
	if err != nil {
		return err
	}

	currentEnvVars, err := loadEnvVars(envFilePath)
	if err != nil {
		return err
	}

	changes := determineEnvVarChanges(previousEnvVars, currentEnvVars)

	statusFilePtr, err := openFile(statusFilePath)
	if err != nil {
		return err
	}
	defer statusFilePtr.Close()

	stagingFilePtr, err := openFileAppendMode(stagingAreaFilePath)
	if err != nil {
		return err
	}
	defer stagingFilePtr.Close()

	for _, change := range changes {
		lineStr := fmt.Sprintf("%s=%s %s\n", change.Key, change.Value, change.Modification)
		_, _ = statusFilePtr.WriteString(lineStr)
		if change.Modification != "unchanged" {
			_, _ = stagingFilePtr.WriteString(lineStr)
		}
	}

	return nil
}

func loadEnvVars(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	envVars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envVars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return envVars, nil
}

func determineEnvVarChanges(previous, current map[string]string) []envVarMetadata {
	changes := []envVarMetadata{}

	for key, currValue := range current {
		if prevValue, exists := previous[key]; exists {
			if currValue != prevValue {
				changes = append(changes, envVarMetadata{Key: key, Value: currValue, Modification: "modified"})
			} else {
				changes = append(changes, envVarMetadata{Key: key, Value: currValue, Modification: "unchanged"})
			}
		} else {
			changes = append(changes, envVarMetadata{Key: key, Value: currValue, Modification: "added"})
		}
	}

	for key, prevValue := range previous {
		if _, exists := current[key]; !exists {
			changes = append(changes, envVarMetadata{Key: key, Value: prevValue, Modification: "removed"})
		}
	}

	return changes
}
