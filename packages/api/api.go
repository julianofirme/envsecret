package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type Secret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetOrganizationsResponse struct {
	Organizations []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"organizations"`
}

const USER_AGENT = "cli"
const ENVSECRET_URL = "http://localhost:3000"

type GetWorkSpacesResponse struct {
	Workspaces []struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		Slug           string `json:"slug"`
		AvatarUrl      string `json:"avatar_url"`
		CreatedAt      string `json:"created_at"`
		UpdatedAt      string `json:"updated_at"`
		OwnerId        string `json:"ownerId"`
		OrganizationId string `json:"organization_id"`
	} `json:"projects"`
}

type GetSecretsResponse struct {
	Secret []struct {
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
	} `json:"secrets"`
}

type SelectOrganizationRequest struct {
	OrganizationId string `json:"organizationId"`
}

type SelectOrganizationResponse struct {
	Token string `json:"token"`
}

func CallLogin(email, password string) (string, error) {
	loginData := map[string]string{"email": email, "password": password}
	body, _ := json.Marshal(loginData)
	req, err := http.NewRequest("POST", "http://localhost:3000/api/users/login", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("login failed: %s", resp.Status)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["token"], nil
}

func CallGetAllOrganizations(httpClient *resty.Client) (GetOrganizationsResponse, error) {
	var orgResponse GetOrganizationsResponse
	response, err := httpClient.
		R().
		SetResult(&orgResponse).
		SetHeader("User-Agent", USER_AGENT).
		Get(fmt.Sprintf("%v/api/organization/cli", ENVSECRET_URL))

	if err != nil {
		return GetOrganizationsResponse{}, err
	}

	if response.IsError() {
		return GetOrganizationsResponse{}, fmt.Errorf("CallGetAllOrganizations: Unsuccessful response: [response=%v]", response)
	}

	return orgResponse, nil
}

func CallSelectOrganization(httpClient *resty.Client, request SelectOrganizationRequest) (SelectOrganizationResponse, error) {
	var selectOrgResponse SelectOrganizationResponse

	response, err := httpClient.
		R().
		SetBody(request).
		SetResult(&selectOrgResponse).
		SetHeader("User-Agent", USER_AGENT).
		Post(fmt.Sprintf("%v/api/organization/select", ENVSECRET_URL))

	if err != nil {
		return SelectOrganizationResponse{}, err
	}

	if response.IsError() {
		return SelectOrganizationResponse{}, fmt.Errorf("CallSelectOrganization: Unsuccessful response: [response=%v]", response)
	}

	return selectOrgResponse, nil

}

func CallGetAllWorkSpacesUserBelongsTo(httpClient *resty.Client, orgId string) (GetWorkSpacesResponse, error) {
	var workSpacesResponse GetWorkSpacesResponse
	response, err := httpClient.
		R().
		SetResult(&workSpacesResponse).
		SetHeader("User-Agent", USER_AGENT).
		Get(fmt.Sprintf("%v/api/project/%s", ENVSECRET_URL, orgId))

	if err != nil {
		return GetWorkSpacesResponse{}, err
	}

	if response.IsError() {
		return GetWorkSpacesResponse{}, fmt.Errorf("CallGetAllWorkSpacesUserBelongsTo: Unsuccessful response:  [response=%v]", response)
	}

	return workSpacesResponse, nil
}

func CallGetSecrets(httpClient *resty.Client, workspaceId string) (GetSecretsResponse, error) {
	var secretsResponse GetSecretsResponse
	response, err := httpClient.
		R().
		SetResult(&secretsResponse).
		SetHeader("User-Agent", USER_AGENT).
		Get(fmt.Sprintf("%v/api/secret/%s", ENVSECRET_URL, workspaceId))

	if err != nil {
		return GetSecretsResponse{}, err
	}

	if response.IsError() {
		return GetSecretsResponse{}, fmt.Errorf("CallGetSecrets: Unsuccessful response:  [response=%v]", response)
	}

	return secretsResponse, nil
}
