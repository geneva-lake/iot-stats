package config

import (
	"encoding/json"
	"io/ioutil"
)

type Mongo struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Config struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	Expiration int    `json:"expiration"`
	Mongo      Mongo  `json:"mongo"`
	ApiKey     string `json:"api-key"`
	Login      string `json:"login"`
	Password   string `json:"password"`
}

func Configuration(configFile string) (*Config, error) {
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
