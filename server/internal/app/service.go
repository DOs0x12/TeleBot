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

	handleServices(ctx, botConf, msgBroker, fromBrokerDataChan, fromBotDataChan)

	return nil
}

func handleServices(ctx context.Context,
	botConf BotConf,
	msgBroker brokerInterf.MessageBroker,
	fromBrokerDataChan <-chan broker.DataFrom,
	fromBotDataChan <-chan botEnt.Data) {
	for {
		select {
		case <-ctx.Done():
			stopServices(botConf, msgBroker)

			return
		case fromBrokerData := <-fromBrokerDataChan:
			go processFromBrokerData(ctx, fromBrokerData, botConf, msgBroker)
		case fromBotData := <-fromBotDataChan:
			go processFromBotData(ctx, fromBotData, botConf, msgBroker)
		}
	}
}

func stopServices(botConf BotConf, msgBroker brokerInterf.MessageBroker) {
	botConf.BotWorker.Stop()
	logrus.Info("The bot has been stopped")
	msgBroker.Close()
	logrus.Info("The transmitter connection has been closed")
	botConf.Storage.Close()
	logrus.Info("The storage connection has been closed")
}

func processFromBrokerData(ctx context.Context,
	fromBrokerData broker.DataFrom,
	botConf BotConf,
	msgBroker brokerInterf.MessageBroker) {
	if fromBrokerData.IsCommand {
		comm, err := unmarshalBotCommand(fromBrokerData.Value)
		if err != nil {
			logrus.Error("Failed to unmarshal a bot command: ", err)

			return
		}

		err = registerBotCommand(ctx, comm, botConf)
		if err != nil {
			logrus.Error("Failed to register a command in the bot: ", err)

			return
		}

		return
	}

	toBotData, err := unmarshalBotData(fromBrokerData.Value)
	if err != nil {
		logrus.Error("Failed to unmarshal bot data: ", err)
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

func toBrokerData(fromBotData botEnt.Data, botConf BotConf) (broker.DataTo, error) {
	if !fromBotData.IsCommand {
		return broker.DataTo{ChatID: fromBotData.ChatID, Value: fromBotData.Value}, nil
	}

	botCommand, err := searchBotCommandByName(fromBotData.Value, *botConf.BotCommands)
	if err != nil {
		return broker.DataTo{}, err
	}

	return broker.DataTo{
			CommName: botCommand.Name,
			ChatID:   fromBotData.ChatID,
			Value:    fromBotData.Value,
			Token:    botCommand.Token,
		},
		nil
}

func processFromBotData(ctx context.Context,
	fromBotData botEnt.Data,
	botConf BotConf,
	msgBroker brokerInterf.MessageBroker) {
	toBrokerData, err := toBrokerData(fromBotData, botConf)
	if err != nil {
		logrus.Error("Failed to get a bot command: ", err)

		return
	}

	err = msgBroker.TransmitData(ctx, toBrokerData)
	if err != nil {
		logrus.Error("Failed to transmit data to the broker: ", err)
	}
}
