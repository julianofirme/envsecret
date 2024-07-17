package util

import (
	"encoding/json"
	"envsecret/packages/models"
	"fmt"
	"io/ioutil"
	"os"
)

func GetFullConfigFilePath() (fullPathToFile string, fullPathToDirectory string, err error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", "", err
	}

	fullPath := fmt.Sprintf("%s/%s/%s", homeDir, CONFIG_FOLDER_NAME, CONFIG_FILE_NAME)
	fullDirPath := fmt.Sprintf("%s/%s", homeDir, CONFIG_FOLDER_NAME)
	return fullPath, fullDirPath, err
}

func GetConfigFile() (models.ConfigFile, error) {
	fullConfigFilePath, _, err := GetFullConfigFilePath()
	if err != nil {
		return models.ConfigFile{}, err
	}

	configFileAsBytes, err := os.ReadFile(fullConfigFilePath)
	if err != nil {
		if err, ok := err.(*os.PathError); ok {
			return models.ConfigFile{}, nil
		} else {
			return models.ConfigFile{}, err
		}
	}

	var configFile models.ConfigFile
	err = json.Unmarshal(configFileAsBytes, &configFile)
	if err != nil {
		return models.ConfigFile{}, err
	}

	return configFile, nil
}

const (
	tokenFile           = "token.txt"
	selectedProjectFile = "selected_project.txt"
)

func SaveToken(token string) error {
	return ioutil.WriteFile(tokenFile, []byte(token), 0644)
}

func LoadToken() (string, error) {
	data, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func SaveSelectedProject(projectId string) error {
	return ioutil.WriteFile(selectedProjectFile, []byte(projectId), 0644)
}

func LoadSelectedProject() (string, error) {
	data, err := ioutil.ReadFile(selectedProjectFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
