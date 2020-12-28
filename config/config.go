package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	myConfig.MinWait = 1000
	myConfig.MaxWait = 1000
	myConfig.StartMinWait = 700
	myConfig.StartMaxWait = 700
	myConfig.Port = 8080
	myConfig.UserCount = 1
	myConfig.NextRandom = true
	myConfig.SelfDataName = "SelfData"
	myConfig.Reset = false
	return myConfig
}

func New(configPath string) *Config {
	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	c, err := Json2Config(dat)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return c
}

func Json2Config(b []byte) (*Config, error) {
	// b := []byte(str)
	c := &Config{}
	err := json.Unmarshal(b, &c)
	return c, err
}

func (config *Config) ToJson() string {
	v, err := json.Marshal(config)
	if err == nil {
		return string(v)
	} else {
		return fmt.Sprintf(`{"error":"%v"}`, err.Error())
	}
}
