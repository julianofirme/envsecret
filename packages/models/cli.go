package models

type UserCredentials struct {
	Email    string `json:"email"`
	JWTToken string `json:"JWTToken"`
}

// The file struct for Envsecret config file
type ConfigFile struct {
	LoggedInUserEmail  string         `json:"loggedInUserEmail"`
	LoggedInUserDomain string         `json:"LoggedInUserDomain,omitempty"`
	LoggedInUsers      []LoggedInUser `json:"loggedInUsers,omitempty"`
	VaultBackendType   string         `json:"vaultBackendType,omitempty"`
}

type LoggedInUser struct {
	Email  string `json:"email"`
	Domain string `json:"domain"`
}
