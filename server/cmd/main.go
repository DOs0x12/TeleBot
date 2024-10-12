package main

import (
	botApp "TeleBot/internal/app/bot"
	"TeleBot/internal/app/interruption"
	botEnt "TeleBot/internal/entities/bot"
	botInfra "TeleBot/internal/infrastructure/bot"
	"TeleBot/internal/infrastructure/config"
	serviceInfra "TeleBot/internal/infrastructure/service"
	"context"
	"flag"

	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("conf", "../etc/config.yml", "Config path.")
	flag.Parse()

	appCtx, appCancel := context.WithCancel(context.Background())

	logrus.Infoln("Start the application")

	interruption.WatchForInterruption(appCancel)

	configer := config.NewConfiger(*configPath)
	config, err := configer.LoadConfig()
	if err != nil {
		logrus.Errorf("Can not get the config data: %v", err)

		return
	}

	commands := &[]botEnt.Command{}
	bot, err := botInfra.NewTelebot(config.BotKey, commands)
	if err != nil {
		logrus.Errorf("A bot error occurs: %v", err)

		return
	}

	receiver := serviceInfra.NewKafkaReceiver("localhost", "9092")
	transmitter := serviceInfra.NewKafkaTransmitter("localhost", "9092")

	err = botApp.Process(appCtx, bot, receiver, transmitter, commands)
	if err != nil {
		logrus.Errorf("An error occurs: %v", err)

		return
	}
}
