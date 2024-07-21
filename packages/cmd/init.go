package cmd

import (
	"encoding/json"
	"envsecret/packages/api"
	"envsecret/packages/models"
	"envsecret/packages/util"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"github.com/manifoldco/promptui"
	"github.com/rs/zerolog/log"
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
	Use:                   "init",
	Short:                 "Used to connect your local project with Envsecret project",
	DisableFlagsInUseLine: true,
	Example:               "es init",
	Args:                  cobra.ExactArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		util.RequireLogin()
	},
	Run: func(_ *cobra.Command, _ []string) {
		if util.WorkspaceConfigFileExistsInCurrentPath() {
			shouldOverride, err := shouldOverrideWorkspacePrompt()
			if err != nil {
				log.Error().Msg("Unable to parse your answer")
				log.Debug().Err(err)
				return
			}

			if !shouldOverride {
				return
			}
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

		organizationResponse, err := api.CallGetAllOrganizations(httpClient)
		if err != nil {
			util.HandleError(err, "Unable to pull organizations that belong to you")
		}

		organizations := organizationResponse.Organizations

		organizationNames := GetOrganizationsNameList(organizationResponse)

		prompt := promptui.Select{
			Label: "Which Envsecret organization would you like to select a project from?",
			Items: organizationNames,
			Size:  7,
		}

		index, _, err := prompt.Run()
		if err != nil {
			util.HandleError(err)
		}

		selectedOrganization := organizations[index]

		tokenResponse, err := api.CallSelectOrganization(httpClient, api.SelectOrganizationRequest{OrganizationId: selectedOrganization.ID})

		if err != nil {
			util.HandleError(err, "Unable to select organization")
		}

		// set the config jwt token to the new token
		userCreds.UserCredentials.JWTToken = tokenResponse.Token
		err = util.StoreUserCredsInKeyRing(&userCreds.UserCredentials)
		httpClient.SetAuthToken(tokenResponse.Token)

		if err != nil {
			util.HandleError(err, "Unable to store your user credentials")
		}

		workspaceResponse, err := api.CallGetAllWorkSpacesUserBelongsTo(httpClient, selectedOrganization.ID)
		if err != nil {
			util.HandleError(err, "Unable to pull projects that belong to you")
		}

		filteredWorkspaces, workspaceNames := GetWorkspacesInOrganization(workspaceResponse, selectedOrganization.ID)

		prompt = promptui.Select{
			Label: "Which of your Envsecret projects would you like to connect this project to?",
			Items: workspaceNames,
			Size:  7,
		}

		index, _, err = prompt.Run()
		if err != nil {
			util.HandleError(err)
		}

		err = writeWorkspaceFile(filteredWorkspaces[index])
		if err != nil {
			util.HandleError(err)
		}
	},
}

func writeWorkspaceFile(selectedWorkspace models.Workspace) error {
	workspaceFileToSave := models.WorkspaceConfigFile{
		WorkspaceId: selectedWorkspace.ID,
	}

	marshalledWorkspaceFile, err := json.MarshalIndent(workspaceFileToSave, "", "    ")
	if err != nil {
		return err
	}

	err = util.WriteToFile(util.ENVSECRET_WORKSPACE_CONFIG_FILE_NAME, marshalledWorkspaceFile, 0600)
	if err != nil {
		return err
	}

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

func shouldOverrideWorkspacePrompt() (bool, error) {
	prompt := promptui.Select{
		Label: "A workspace config file already exists here. Would you like to override? Select[Yes/No]",
		Items: []string{"No", "Yes"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return result == "Yes", nil
}

func GetOrganizationsNameList(organizationResponse api.GetOrganizationsResponse) []string {
	organizations := organizationResponse.Organizations

	if len(organizations) == 0 {
		message := fmt.Sprintf("You don't have any organization created in envsecret. You must first create a organization at %s", util.ENVSECRET_DEFAULT_URL)
		util.PrintErrorMessageAndExit(message)
	}

	var organizationNames []string
	for _, workspace := range organizations {
		organizationNames = append(organizationNames, workspace.Name)
	}

	return organizationNames
}

func GetWorkspacesInOrganization(workspaceResponse api.GetWorkSpacesResponse, orgId string) ([]models.Workspace, []string) {
	workspaces := workspaceResponse.Workspaces

	var filteredWorkspaces []models.Workspace
	var workspaceNames []string

	for _, workspace := range workspaces {
		if workspace.OrganizationId == orgId {
			filteredWorkspaces = append(filteredWorkspaces, workspace)
			workspaceNames = append(workspaceNames, workspace.Name)
		}
	}

	if len(filteredWorkspaces) == 0 {
		message := fmt.Sprintf("You don't have any projects created in Envsecret organization. You must first create a project at %s", util.ENVSECRET_DEFAULT_URL)
		util.PrintErrorMessageAndExit(message)
	}

	return filteredWorkspaces, workspaceNames
}
