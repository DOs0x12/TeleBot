package main

import (
	botApp "TeleBot/internal/app/bot"
	botCommCreator "TeleBot/internal/app/bot/command"
	"TeleBot/internal/app/interruption"
	botEnt "TeleBot/internal/entities/bot"
	botInfra "TeleBot/internal/infrastructure/bot"
	"TeleBot/internal/infrastructure/config"
	serviceInfra "TeleBot/internal/infrastructure/service"
	"context"

	"github.com/sirupsen/logrus"
)

const configPath = "../etc/config.yml"

func main() {
	appCtx, appCancel := context.WithCancel(context.Background())

	logrus.Infoln("Start the application")

	interruption.WatchForInterruption(appCancel)

	configer := config.NewConfiger(configPath)
	helloComm := botCommCreator.CreateHelloCommand()
	commands := &[]botEnt.Command{helloComm}
	bot, err := botInfra.NewTelebot(configer, commands)
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
