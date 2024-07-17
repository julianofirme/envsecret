package cmd

import (
	"envsecret/packages/util"
	"fmt"

	"github.com/spf13/cobra"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to the API",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("e")
		password, _ := cmd.Flags().GetString("p")

		token, err := util.Login(username, password)
		if err != nil {
			fmt.Println("Login failed:", err)
			return
		}

		err = util.SaveToken(token)
		if err != nil {
			fmt.Println("Failed to save token:", err)
			return
		}

		fmt.Println("Login successful, token saved.")
	},
}

func init() {
	LoginCmd.Flags().String("e", "", "Email for login")
	LoginCmd.Flags().String("p", "", "Password for login")
	LoginCmd.MarkFlagRequired("e")
	LoginCmd.MarkFlagRequired("p")

	rootCmd.AddCommand(LoginCmd)
}
