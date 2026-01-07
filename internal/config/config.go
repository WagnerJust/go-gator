package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}
const configFileName = ".gatorconfig.json"


func NewConfig() *Config {
	return &Config{}
}

func (c *Config) String() string {
 data, err := json.MarshalIndent(c, "", "  ")
 if err != nil {
  return ""
 }
 return string(data)
}

func getConfigFilePath() (string, error) {
	path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path = path + "/" + configFileName
	return path, nil
}

func (c *Config) Read () (error) {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

func (c *Config) SetUser (user string) (error) {
	c.CurrentUserName = user
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	defer file.Close()

	if err != nil {
		return err
	}
	return nil
}
