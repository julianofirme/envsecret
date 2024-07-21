package cmd

import (
	"encoding/json"
	"envsecret/packages/api"
	"envsecret/packages/models"
	"envsecret/packages/util"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets for your project",
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull secrets from the API and create/update .env file",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig(".envsecret.json")
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		userCreds, err := util.GetCurrentLoggedInUserDetails()
		if err != nil {
			util.HandleError(err, "Unable to get your login details")
		}

		if userCreds.LoginExpired {
			util.PrintErrorMessageAndExit("Your login session has expired, please run [envsecret login] and try again")
		}

		httpClient := resty.New()
		httpClient.SetAuthToken(userCreds.UserCredentials.JWTToken)

		secrets, err := api.CallGetSecrets(httpClient, config.WorkspaceId)
		if err != nil {
			fmt.Println("Error parsing secrets:", err)
			return
		}

		file, err := os.OpenFile(".env", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Println("Error creating .env file:", err)
			return
		}
		defer file.Close()

		for _, secret := range secrets.Secret {
			_, err := file.WriteString(fmt.Sprintf("%s=%s\n", secret.Key, secret.Value))
			if err != nil {
				fmt.Println("Error writing to .env file:", err)
				return
			}
		}

		fmt.Println(".env file created/updated successfully.")
	},
}

func loadConfig(filename string) (*models.WorkspaceConfigFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config models.WorkspaceConfigFile
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func init() {
	secretCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(secretCmd)
}
