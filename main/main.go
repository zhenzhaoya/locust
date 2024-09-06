package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/zhenzhaoya/locust"
	"github.com/zhenzhaoya/locust/utils"
)

var (
	app *locust.Locust
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "config.json", "config file path")
	flag.Parse()
	log.Println("config file: " + configFile)
	app = locust.GetAPP(utils.GetPath("../index.html"), utils.GetPath("../http"))
	c := locust.SetConfig(app, configFile)
	if c != nil {
		log.Println("port: " + strconv.Itoa(c.Port))
		app.Start(c.Port)
	} else {
		log.Println("port: 8080")
		app.Start(8080)
	}

	log.Println("app exited...")
}

// go run ./main/ --config config.toml
