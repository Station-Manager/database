package main

import (
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/database"
	"github.com/Station-Manager/logging"
	"path/filepath"
)

func main() {
	fp, err := filepath.Abs("../build")
	if err != nil {
		panic(err)
	}
	cfgService := &config.Service{
		WorkingDir: fp,
		//		AppConfig:  cfg,
	}
	if err = cfgService.Initialize(); err != nil {
		panic(err)
	}

	loggingService := &logging.Service{ConfigService: cfgService}
	if err = loggingService.Initialize(); err != nil {
		panic(err)
	}
	defer func() { _ = loggingService.Close() }()

	dbService := database.Service{ConfigService: cfgService, Logger: loggingService}
	if err = dbService.Initialize(); err != nil {
		panic(err)
	}

	if err = dbService.Open(); err != nil {
		panic(err)
	}

	if err = dbService.Migrate(); err != nil {
		panic(err)
	}

	if err = dbService.Close(); err != nil {
		panic(err)
	}
}
