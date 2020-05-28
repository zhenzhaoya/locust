package config

import (
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	BaseUrl      string
	HttpFile     string
	Port         int
	UserCount    int
	MinWait      int
	MaxWait      int
	SelfDataName string
}

func New(configPath string) *Config {
	myConfig := &Config{}
	myConfig.loadConfig(configPath)
	return myConfig
}

func (config *Config) loadConfig(configPath string) {
	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if _, err := toml.Decode(string(dat), config); err != nil {
		log.Fatalf("error: %v", err)
	}
}
