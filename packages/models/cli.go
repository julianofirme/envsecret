package models

type UserCredentials struct {
	Email        string `json:"email"`
	PrivateKey   string `json:"privateKey"`
	JTWToken     string `json:"JTWToken"`
	RefreshToken string `json:"RefreshToken"`
}

// The file struct for Infisical config file
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
