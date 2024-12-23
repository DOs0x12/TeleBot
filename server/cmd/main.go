package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	botApp "github.com/DOs0x12/TeleBot/server/internal/app"
	botEnt "github.com/DOs0x12/TeleBot/server/internal/entities/bot"
	botInfra "github.com/DOs0x12/TeleBot/server/internal/infrastructure/bot"
	brokerInfra "github.com/DOs0x12/TeleBot/server/internal/infrastructure/broker"
	"github.com/DOs0x12/TeleBot/server/internal/infrastructure/config"
	"github.com/DOs0x12/TeleBot/server/internal/infrastructure/storage"

	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("conf", "../etc/config.yml", "Config path.")
	storageAddress := flag.String("stAddr", "localhost", "The address of the storage.")
	storageDB := flag.String("stDB", "telebot", "The name of the storage database.")
	storageUser := flag.String("stUser", "user", "The user of the storage database.")
	storagePass := flag.String("stPass", "", "The password of the storage user.")
	flag.Parse()

	logrus.Info("Load the application data")

	if *storagePass == "" {
		logrus.Error("The password for the storage user is not passed")
		flag.PrintDefaults()

		return
	}

	pgStorage, err := storage.NewPgCommStorage(*storageAddress, *storageDB, *storageUser, *storagePass)
	if err != nil {
		logrus.Error("Can not create a storage: ", err)

		return
	}

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

	cons, err := brokerInfra.NewKafkaConsumer(config.KafkaAddress)
	if err != nil {
		logrus.Error("A receiver creating error occurs:", err)
	}

	prod := brokerInfra.NewKafkaProducer(config.KafkaAddress)

	logrus.Info("Start the application")

	err = botApp.Process(appCtx, bot, cons, prod, &commands, pgStorage)
	if err != nil {
		logrus.Error("An error of the application work occurs:", err)

		return
	}
}
