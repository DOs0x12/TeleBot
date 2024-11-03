package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	botApp "github.com/Guise322/TeleBot/server/internal/app"
	botEnt "github.com/Guise322/TeleBot/server/internal/entities/bot"
	botInfra "github.com/Guise322/TeleBot/server/internal/infrastructure/bot"
	serviceInfra "github.com/Guise322/TeleBot/server/internal/infrastructure/broker"
	"github.com/Guise322/TeleBot/server/internal/infrastructure/config"

	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("conf", "../etc/config.yml", "Config path.")
	flag.Parse()

	logrus.Infoln("Load the application configuration")

	appCtx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	configer := config.NewConfiger(*configPath)
	config, err := configer.LoadConfig()
	if err != nil {
		logrus.Errorf("Can not get the config data: %v", err)

		return
	}

	commands := []botEnt.Command{}
	bot, err := botInfra.NewTelebot(config.BotKey, commands)
	if err != nil {
		logrus.Errorf("A bot loading error occurs: %v", err)

		return
	}

	receiver := serviceInfra.NewKafkaReceiver(config.KafkaAddress)
	transmitter := serviceInfra.NewKafkaTransmitter(config.KafkaAddress)

	logrus.Infoln("Start the application")

	err = botApp.Process(appCtx, bot, receiver, transmitter, &commands)
	if err != nil {
		logrus.Errorf("An error occurs: %v", err)

		return
	}
}
