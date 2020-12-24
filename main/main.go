package main

import (
	"flag"
	"log"

	"github.com/zhenzhaoya/locust"
	"github.com/zhenzhaoya/locust/utils"
)

var (
	app *locust.Locust //= locust.GetAPP("../index.html")
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "config.json", "config file path")
	flag.Parse()
	log.Println("config file: " + configFile)
	app = locust.GetAPP(utils.GetPath("../index.html"), utils.GetPath("../http"))
	c := locust.SetConfig(app, configFile)
	if c != nil {
		app.Start(c.Port)
	} else {
		app.Start(8080)
	}

	log.Println("app exited...")
}

// go run ./main/ --config config.toml
