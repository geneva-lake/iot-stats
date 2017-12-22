package main

import (
	"flag"
	"iot-stats/config"
	"iot-stats/model"
	"iot-stats/server"
	"iot-stats/service"
	"iot-stats/utils"
	"os"
)

const defaultConfigFile = "config.json"

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", defaultConfigFile, "Config file")
	flag.Parse()
	os.Exit(run(configFile))
}

func run(configFile string) int {
	cfg, err := config.Configuration(configFile)
	if err != nil {
		utils.Log().Infoln("run error", err)
		return 1
	}
	ms := service.NewMongoService(&service.Config{
		Host:     cfg.Mongo.Host,
		Port:     cfg.Mongo.Port,
		User:     cfg.Mongo.User,
		Password: cfg.Mongo.Password,
		Database: cfg.Mongo.Database,
	})
	if err = ms.Connect(); err != nil {
		utils.Log().Infoln("mongo connection error", err)
		return 1
	}
	err = bootstrap(cfg.Login, cfg.Password, ms)
	if err != nil {
		utils.Log().Infoln("run error", err)
		return 1
	}
	srv := server.NewServer(&server.Config{
		Host:       cfg.Host,
		Port:       cfg.Port,
		ApiKey:     cfg.ApiKey,
		Expiration: cfg.Expiration,
	}, ms)
	if err := srv.Serve(); err != nil {
		utils.Log().Infoln("run error", err)
		return 1
	}
	return 0
}

func bootstrap(login, password string, ms *service.MongoService) error {
	creds := model.Credentials{
		Login:    login,
		Password: utils.GenerateHash(password),
	}
	if err := ms.SetCreds(creds); err != nil {
		return err
	} else {
		return nil
	}
}
