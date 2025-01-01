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
	flag.Parse()

	appCtx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	logrus.Info("Load the application configuration")

	configer := config.NewConfiger(*configPath)
	config, err := configer.LoadConfig()
	if err != nil {
		logrus.Error("Failed to get the config data: ", err)

		return
	}

	storageConf := storage.StorageConf{Address: config.StorageAddress,
		Database: config.StorageDB,
		User:     config.StorageUser,
		Pass:     config.StoragePass,
	}

	pgStorage, err := storage.NewPgCommStorage(appCtx, storageConf)
	if err != nil {
		logrus.Error("Failed to create a storage: ", err)

		return
	}

	commands := []botEnt.Command{}
	bot, err := botInfra.NewTelebot(appCtx, config.BotKey, commands)
	if err != nil {
		logrus.Error("Failed to start up a bot: ", err)

		return
	}

	botConf := botApp.BotConf{BotWorker: bot, BotCommands: &commands, Storage: pgStorage}

	cons, err := brokerInfra.NewKafkaConsumer(config.KafkaAddress)
	if err != nil {
		logrus.Error("Failed to create a receiver: ", err)
	}

	prod := brokerInfra.NewKafkaProducer(config.KafkaAddress)

	brokerConf := botApp.BrokerConf{Receiver: cons, Transmitter: prod}

	logrus.Info("Start the application")

	err = botApp.Process(appCtx, botConf, brokerConf)
	if err != nil {
		logrus.Error("An application error occured: ", err)

		return
	}
}
