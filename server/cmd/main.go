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

	logrus.Info("Load the application configuration")

	appCtx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	configer := config.NewConfiger(*configPath)
	config, err := configer.LoadConfig()
	if err != nil {
		logrus.Error("Can not get the config data:", err)

		return
	}

	commands := []botEnt.Command{}
	bot, err := botInfra.NewTelebot(appCtx, config.BotKey, commands)
	if err != nil {
		logrus.Error("A bot loading error occurs:", err)

		return
	}

	receiver := serviceInfra.NewKafkaReceiver(config.KafkaAddress)
	transmitter := serviceInfra.NewKafkaTransmitter(config.KafkaAddress)

	logrus.Info("Start the application")

	err = botApp.Process(appCtx, bot, receiver, transmitter, &commands)
	if err != nil {
		logrus.Error("An error occurs:", err)

		return
	}
}
