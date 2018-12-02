package process

import (
	"log"

	"../models"
	"../utils"

	"github.com/BurntSushi/toml"
)

// Init is init data
func Init(runTime, confPath string) {
	// init runtime
	if _, err := toml.DecodeFile(confPath, &models.RunTimeMap); err != nil {
		log.Println(err)
		return
	}
	models.RunTimeInfo = models.RunTimeMap[runTime]
	log.Println(models.RunTimeInfo)
	// init log
	util.InitLog()
	// init IP
	util.GetIP()
	// init channel
	models.OperateChan = make(chan models.OperateModel, 100)
}
