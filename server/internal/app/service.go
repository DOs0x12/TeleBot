package telebot

import (
	"context"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	"github.com/DOs0x12/TeleBot/server/v2/internal/entities/broker"
	botInterf "github.com/DOs0x12/TeleBot/server/v2/internal/interfaces/bot"
	brokerInterf "github.com/DOs0x12/TeleBot/server/v2/internal/interfaces/broker"
	storInterf "github.com/DOs0x12/TeleBot/server/v2/internal/interfaces/storage"

	"github.com/sirupsen/logrus"
)

type BotDataDto struct {
	ChatID int64
	Value  string
}

type BotCommandDto struct {
	Name        string
	Description string
}

type BotConf struct {
	BotWorker   botInterf.Worker
	BotCommands *[]botEnt.Command
	Storage     storInterf.CommandStorage
}

func Process(ctx context.Context,
	botConf BotConf,
	msgBroker brokerInterf.MessageBroker,
) error {
	fromBrokerDataChan, err := msgBroker.StartReceivingData(ctx)
	if err != nil {
		return fmt.Errorf("failed to start receiving data: %w", err)
	}

	fromBotDataChan := botConf.BotWorker.Start(ctx)
	if err = loadBotCommands(ctx, botConf.BotCommands, botConf.Storage); err != nil {
		return err
	}

	if err = botConf.BotWorker.RegisterCommands(ctx, *botConf.BotCommands); err != nil {
		return fmt.Errorf("failed to register the loaded commands: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			botConf.BotWorker.Stop()
			logrus.Info("The bot has been stopped")
			msgBroker.Close()
			logrus.Info("The transmitter connection has been closed")
			botConf.Storage.Close()
			logrus.Info("The storage connection has been closed")

			return nil
		case fromBrokerData := <-fromBrokerDataChan:
			go processFromBrokerData(ctx, fromBrokerData, botConf, msgBroker)
		case fromBotData := <-fromBotDataChan:
			toBrokerData, err := processFromBotData(fromBotData, *botConf.BotCommands)
			if err != nil {
				logrus.Error("Failed to process data from the bot: ", err)

				continue
			}

			go func() {
				err = msgBroker.TransmitData(ctx, toBrokerData)
				if err != nil {
					logrus.Error("Failed to transmit data to the broker: ", err)
				}
			}()
		}
	}
}

func processFromBrokerData(ctx context.Context,
	fromBrokerData broker.DataFrom,
	botConf BotConf,
	msgBroker brokerInterf.MessageBroker) {
	if fromBrokerData.IsCommand {
		comm, err := command(fromBrokerData.Value)
		if err != nil {
			logrus.Error("Failed to get a command form a broker message: ", err)

			return
		}

		err = registerBotCommand(ctx, comm, botConf)
		if err != nil {
			logrus.Error("Failed to register a command in the bot: ", err)

			return
		}

		return
	}

	toBotData, err := castFromBrokerData(fromBrokerData)
	if err != nil {
		logrus.Error("Failed to cast data for sending it to the bot: ", err)
	}

	err = botConf.BotWorker.SendMessage(ctx, toBotData.Value, toBotData.ChatID)
	if err != nil {
		logrus.Error("Dailed to send a message to the bot: ", err)

		return
	}

	err = msgBroker.Commit(ctx, fromBrokerData.MsgUuid)
	if err != nil {
		logrus.WithField("messageUuid", err).Error("Failed to commit the message with UUID: ", err)
	}
}
