package cmd

import (
	"bufio"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

const (
	esTimeFormat = "2006-01-02 15:04:05"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "This allows you to display all tracked files status",
	Example: "es status",
	RunE: func(_ *cobra.Command, _ []string) error {
		return runStatusCommand(os.Stdout, stagingAreaFilePath)
	},
}

func runStatusCommand(writer io.Writer, stagingAreaFilePath string) error {
	allMetadata, err := getFileMetadataFromStagingFile(stagingAreaFilePath)
	if err != nil {
		return err
	}

	displayStatus(writer, allMetadata)

	return nil
}

// Display Format: | secret | status | last modification time |
func displayStatus(writer io.Writer, allMetadata []fileMetadata) {
	if len(allMetadata) == 0 {
		color.Green("No changes on staging area!")
		return
	}

	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"File name", "Status", "Last Modification Time"})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor},
	)

	for _, mt := range allMetadata {
		row := []string{mt.Path, string(mt.Status), mt.ModificationTime}
		statusColor := tablewriter.FgGreenColor
		if mt.Status == StatusUpdated {
			statusColor = tablewriter.FgBlueColor
		}
		table.Rich(row, []tablewriter.Colors{{}, {tablewriter.Bold, statusColor}, {}})
	}

	table.Render()
}

func getFileMetadataFromStagingFile(stagingAreaFilePath string) ([]fileMetadata, error) {
	filePtr, err := openFile(stagingAreaFilePath)
	if err != nil {
		return nil, err
	}
	defer filePtr.Close()

	updateLatestStateMap := make(map[string]fileMetadata)

	scanner := bufio.NewScanner(filePtr)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		mData := extractFileMetadataFromLine(line)
		if (mData != fileMetadata{}) {
			updateLatestStateMap[mData.Path] = mData
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var allMetadata []fileMetadata
	for _, v := range updateLatestStateMap {
		allMetadata = append(allMetadata, v)
	}

	sort.Slice(allMetadata, func(i, j int) bool {
		return allMetadata[i].Path < allMetadata[j].Path
	})

	return allMetadata, nil
}

func extractFileMetadataFromLine(lineStr string) fileMetadata {
	// Assuming the format is `VARIAVEL="VALOR" status`
	parts := strings.Fields(lineStr)
	if len(parts) < 2 {
		return fileMetadata{}
	}
	variable := parts[0]
	status := parts[1]

	modTime := time.Now().Format(esTimeFormat) // Placeholder for actual modification time

	return fileMetadata{
		Path:             variable,
		ModificationTime: modTime,
		Status:           FileStatus(status),
	}
}
