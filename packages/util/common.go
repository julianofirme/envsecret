package util

import (
	"fmt"
	"os"
)

func GetHomeDir() (string, error) {
	directory, err := os.UserHomeDir()
	return directory, err
}

func WriteToFile(fileName string, dataToWrite []byte, filePerm os.FileMode) error {
	err := os.WriteFile(fileName, dataToWrite, filePerm)
	if err != nil {
		return fmt.Errorf("unable to wrote to file [err=%v]", err)
	}

	return nil
}
