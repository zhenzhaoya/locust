package config

import (
	"encoding/json"
	"fmt"
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
	StartMinWait int
	StartMaxWait int
	NextRandom   bool
	SelfDataName string
	Reset        bool
	SelfConfig   interface{}
}

func GetDefault() *Config {
	myConfig := &Config{}
	myConfig.MinWait = 5000
	myConfig.MaxWait = 9000
	myConfig.StartMinWait = 1000
	myConfig.StartMaxWait = 1000
	myConfig.Port = 8080
	myConfig.UserCount = 1
	myConfig.NextRandom = true
	myConfig.SelfDataName = "SelfData"
	myConfig.Reset = false
	return myConfig
}

func New(configPath string) *Config {
	myConfig := &Config{}
	myConfig.loadConfig(configPath)
	return myConfig
}

func Json2Config(b []byte) (*Config, error) {
	// b := []byte(str)
	c := &Config{}
	err := json.Unmarshal(b, &c)
	return c, err
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

func (config *Config) ToString() string {
	v, err := json.Marshal(config)
	if err == nil {
		return string(v)
	} else {
		return fmt.Sprintf(`{"error":"%v"}`, err.Error())
	}
}
