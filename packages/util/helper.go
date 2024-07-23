package util

import (
	"strings"
)

func RequireLogin() {
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
	if strings.HasSuffix(address, "/api") {
		return address
	}

	if address[len(address)-1] == '/' {
		return address + "api"
	}
	return address + "/api"
}
