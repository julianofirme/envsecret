package cmd

import (
	"os"
)

func openFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_RDWR, os.ModePerm)
}

func openFileAppendMode(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
}
