package main

import (
	"flag"
	"os"

	"./models"
	"./process"

	_ "net/http/pprof"
)

// RunTime=dev go run app.go -conf "./conf/conf.toml"

var confPath = flag.String("conf", "./conf/conf.toml", "The conf path.")

func main() {
	models.RunTime = os.Getenv("RunTime")
	flag.Parse()
	// init
	process.Init(models.RunTime, *confPath)
	// create server
	process.CreateServer()
}
