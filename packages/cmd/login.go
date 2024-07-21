package cmd

import (
	"envsecret/packages/api"
	"envsecret/packages/models"
	"envsecret/packages/util"
	"errors"
	"fmt"
	"regexp"

	"github.com/go-resty/resty/v2"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var userCredentialsToBeStored models.UserCredentials

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to the API",
	Run: func(cmd *cobra.Command, args []string) {
		email, password, err := askForLoginCredentials()
		if err != nil {
			util.HandleError(err, "Unable to parse email and password for authentication")
		}

		token, err := api.CallLogin(email, password)

		userCredentialsToBeStored.Email = email
		userCredentialsToBeStored.JWTToken = token

		if err != nil {
			fmt.Println("Unable to authenticate with the provided credentials, please try again")
			log.Debug().Err(err)
			//return here
			util.HandleError(err)
		}

		err = util.StoreUserCredsInKeyRing(&userCredentialsToBeStored)
		if err != nil {
			util.HandleError(err, "[envsecret user update domain]: Unable to store credentials")
		}

		configFile, err := util.GetConfigFile()
		if err != nil {
			util.HandleError(err, "[envsecret user update domain]: Unable to get config file")
		}

		configFile.LoggedInUserEmail = email
		configFile.LoggedInUserToken = token

		err = util.WriteConfigFile(&configFile)
		if err != nil {
			log.Error().Msgf("Unable to store your credentials in system vault")
			log.Debug().Err(err)
			//return here
			util.HandleError(err)
		}
		if err != nil {
			fmt.Println("Failed to save token:", err)
			return
		}

		fmt.Println("Login successful, token saved.")
	},
}

func askForLoginCredentials() (email string, password string, err error) {
	validateEmail := func(input string) error {
		matched, err := regexp.MatchString("^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$", input)
		if err != nil || !matched {
			return errors.New("this doesn't look like an email address")
		}
		return nil
	}

	fmt.Println("Enter Credentials...")
	emailPrompt := promptui.Prompt{
		Label:    "Email",
		Validate: validateEmail,
	}

	userEmail, err := emailPrompt.Run()

	if err != nil {
		return "", "", err
	}

	validatePassword := func(input string) error {
		if len(input) < 1 {
			return errors.New("please enter a valid password")
		}
		return nil
	}

	passwordPrompt := promptui.Prompt{
		Label:    "Password",
		Validate: validatePassword,
		Mask:     '*',
	}

	userPassword, err := passwordPrompt.Run()

	if err != nil {
		return "", "", err
	}

	return userEmail, userPassword, nil
}

func GetJwtTokenWithOrganizationId(oldJwtToken string) string {
	log.Debug().Msg(fmt.Sprint("GetJwtTokenWithOrganizationId: ", "oldJwtToken", oldJwtToken))

	httpClient := resty.New()
	httpClient.SetAuthToken(oldJwtToken)

	organizationResponse, err := api.CallGetAllOrganizations(httpClient)

	if err != nil {
		util.HandleError(err, "Unable to pull organizations that belong to you")
	}

	organizations := organizationResponse.Organizations

	organizationNames := GetOrganizationsNameList(organizationResponse)

	prompt := promptui.Select{
		Label: "Which Envsecret organization would you like to log into?",
		Items: organizationNames,
	}

	index, _, err := prompt.Run()
	if err != nil {
		util.HandleError(err)
	}

	selectedOrganization := organizations[index]

	selectedOrgRes, err := api.CallSelectOrganization(httpClient, api.SelectOrganizationRequest{OrganizationId: selectedOrganization.ID})

	if err != nil {
		util.HandleError(err)
	}

	return selectedOrgRes.Token

}

func init() {
	rootCmd.AddCommand(LoginCmd)
}
