package main

import (
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/database"
	"github.com/Station-Manager/types"
)

func main() {
	cfg := types.AppConfig{
		DatastoreConfig: types.DatastoreConfig{
			Driver:                    database.PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "station_manager",
			User:                      "smuser",
			Password:                  "1q2w3e4r",
			SSLMode:                   "disable",
			MaxOpenConns:              1,
			MaxIdleConns:              1,
			ConnMaxLifetime:           1,
			ConnMaxIdleTime:           1,
			ContextTimeout:            5,
			TransactionContextTimeout: 5,
		},
	}
	cfgService := &config.Service{
		WorkingDir: "",
		AppConfig:  cfg,
	}
	if err := cfgService.Initialize(); err != nil {
		panic(err)
	}

	dbService := database.Service{
		ConfigService: cfgService,
	}
	if err := dbService.Initialize(); err != nil {
		panic(err)
	}

	if err := dbService.Open(); err != nil {
		panic(err)
	}

	if err := dbService.Migrate(); err != nil {
		panic(err)
	}

	if err := dbService.Close(); err != nil {
		panic(err)
	}
}
