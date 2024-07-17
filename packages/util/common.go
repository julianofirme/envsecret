package util

import "os"

func GetHomeDir() (string, error) {
	directory, err := os.UserHomeDir()
	return directory, err
}
