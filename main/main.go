package main

import (
	"flag"
	"locust"
	"locust/config"
	"locust/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	app *locust.Locust //= locust.GetAPP("../index.html")
)

func main() {
	// fmt.Println(rootPath)
	var configFile string
	flag.StringVar(&configFile, "config", "config.toml", "config file path")
	flag.Parse()
	log.Println("config file: " + configFile)

	if configFile != "" {
		configFile = getPath(configFile)
		c := config.New(configFile)
		app = locust.GetAPP(getPath("index.html"))
		if c.MinWait > 0 {
			app.MinWait = c.MinWait
		}
		if c.MaxWait > 0 {
			app.MaxWait = c.MaxWait
		}
		app.UserCount = c.UserCount
		app.SelfDataName = c.SelfDataName
		dic := make(map[string]interface{}, 0)
		dic["BaseUrl"] = c.BaseUrl
		b := locust.Init(getPath(c.HttpFile), app, dic)
		if !b {
			return
		}
		if c.Port == 0 {
			c.Port = 8080
		}
		app.Start(c.Port)
	}
	log.Println("app exited...")
}
func getPath(path string) string {
	b, _ := utils.PathExists(path)
	if !b {
		rootPath := filepath.Dir(os.Args[0])
		if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
			path = rootPath + path
		} else {
			path = rootPath + "/" + path
		}
		b, _ = utils.PathExists(path)
		if !b {
			log.Println("file not exist. " + path)
		}
	}
	return path
}

// go run ./main/ --config config.toml
