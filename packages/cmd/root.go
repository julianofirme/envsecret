package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	changeGlobalTimeToUTC()
}

var UseDefaultTime bool

var rootCmd = &cobra.Command{
	Use:   "es",
	Short: "envsecret is a Command Line Interface (CLI) to version control of your .env files.",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&UseDefaultTime, "dtime", false, "use default time (0001-01-01 00:00:00 +0000 UTC)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to execute command. Reason: %#v", err)
	}
}

func changeGlobalTimeToUTC() {
	loc, err := time.LoadLocation("UTC")
	if err == nil {
		time.Local = loc
	}
}
