package models

type UserCredentials struct {
	Email    string `json:"email"`
	JWTToken string `json:"JWTToken"`
}

// The file struct for Envsecret config file
type ConfigFile struct {
	LoggedInUserEmail  string         `json:"loggedInUserEmail"`
	LoggedInUserToken  string         `json:"loggedInUserToken"`
	LoggedInUserDomain string         `json:"LoggedInUserDomain,omitempty"`
	LoggedInUsers      []LoggedInUser `json:"loggedInUsers,omitempty"`
	VaultBackendType   string         `json:"vaultBackendType,omitempty"`
}

type LoggedInUser struct {
	Email  string `json:"email"`
	Domain string `json:"domain"`
}

type WorkspaceConfigFile struct {
	WorkspaceId                   string            `json:"workspaceId"`
	DefaultEnvironment            string            `json:"defaultEnvironment"`
	GitBranchToEnvironmentMapping map[string]string `json:"gitBranchToEnvironmentMapping"`
}
type Workspace struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Slug           string `json:"slug"`
	AvatarUrl      string `json:"avatar_url"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	OwnerId        string `json:"ownerId"`
	OrganizationId string `json:"organization_id"`
}

type Secret struct {
	ID             string `json:"id"`
	Version        int    `json:"version"`
	KeyEncrypted   string `json:"key_encrypted"`
	KeyIV          string `json:"key_iv"`
	KeyAuthTag     string `json:"key_auth_tag"`
	ValueEncrypted string `json:"value_encrypted"`
	ValueIV        string `json:"value_iv"`
	ValueAuthTag   string `json:"value_auth_tag"`
	Algorithm      string `json:"algorithm"`
	KeyEncoding    string `json:"key_encoding"`
	ProjectID      string `json:"project_id"`
	UserID         string `json:"user_id"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	Key            string `json:"key"`
	Value          string `json:"value"`
	Env            string `json:"env"`
}
