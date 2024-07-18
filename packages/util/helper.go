package util

import (
	"strings"
)

func RequireLogin() {
	// get the config file that stores the current logged in user email
	configFile, _ := GetConfigFile()

	if configFile.LoggedInUserEmail == "" {
		PrintErrorMessageAndExit("You must be logged in to run this command. To login, run [es login]")
	}
}

func IsLoggedIn() bool {
	configFile, _ := GetConfigFile()
	return configFile.LoggedInUserEmail != ""
}

func AppendAPIEndpoint(address string) string {
	// Ensure the address does not already end with "/api"
	if strings.HasSuffix(address, "/api") {
		return address
	}

	// Check if the address ends with a slash and append accordingly
	if address[len(address)-1] == '/' {
		return address + "api"
	}
	return address + "/api"
}
