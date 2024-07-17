package util

func RequireLogin() {
	// get the config file that stores the current logged in user email
	token, _ := LoadToken()

	if token == "" {
		PrintErrorMessageAndExit("You must be logged in to run this command. To login, run [es login]")
	}
}
