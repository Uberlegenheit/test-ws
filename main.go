package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
	"websocket-test/api"
	"websocket-test/conf"
	"websocket-test/helpers/modules"
	"websocket-test/service"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	cfg, err := conf.GetNewConfig()
	if err != nil {
		log.Fatal("can`t read config from file ", err)
	}

	s, err := service.NewService()
	if err != nil {
		log.Fatal("services.NewService ", err)
	}

	a, err := api.NewAPI(cfg, s)

	go s.ConnectMarketDataStream()

	if err != nil {
		log.Fatal("api.NewAPI ", err)
	}
	mds := []modules.Module{a}

	modules.Run(mds)

	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	<-gracefulStop
	modules.Stop(mds)
}
