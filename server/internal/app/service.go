package telebot

import (
	"context"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/internal/entities/bot"
	botInterf "github.com/DOs0x12/TeleBot/server/internal/interfaces/bot"
	brokerInterf "github.com/DOs0x12/TeleBot/server/internal/interfaces/broker"
	storInterf "github.com/DOs0x12/TeleBot/server/internal/interfaces/storage"

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

type BrokerConf struct {
	Receiver    brokerInterf.DataReceiver
	Transmitter brokerInterf.DataTransmitter
}

func Process(ctx context.Context,
	botConf BotConf,
	brokerConf BrokerConf,
) error {
	fromBrokerDataChan, err := brokerConf.Receiver.StartReceivingData(ctx)
	if err != nil {
		return fmt.Errorf("an error of the data receiver occurs: %w", err)
	}

	fromBotDataChan := botConf.BotWorker.Start(ctx)
	if err = loadBotCommands(ctx, botConf.BotCommands, botConf.Storage); err != nil {
		return fmt.Errorf("can not load the bot commands: %w", err)
	}

	if err = botConf.BotWorker.RegisterCommands(ctx, *botConf.BotCommands); err != nil {
		return fmt.Errorf("can not register the loaded commands: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			botConf.BotWorker.Stop()
			brokerConf.Transmitter.Close()
			botConf.Storage.Close()
			logrus.Info("The bot is stopped")

			return nil
		case fromBrokerData := <-fromBrokerDataChan:
			go func() {
				err := processFromBrokerData(ctx, fromBrokerData, botConf)
				if err != nil {
					logrus.Error("Can not process data from the broker: ", err)
				}

				err = brokerConf.Receiver.Commit(ctx, fromBrokerData.MsgUuid)
				if err != nil {
					logrus.WithField("messageUuid", err).Error("Can not commit the message with UUID: ", err)
				}
			}()
		case fromBotData := <-fromBotDataChan:
			toBrokerData, err := processFromBotData(fromBotData, *botConf.BotCommands)
			if err != nil {
				logrus.Error("An error of processing bot in data occurs: ", err)

				continue
			}

			go func() {
				err = brokerConf.Transmitter.TransmitData(ctx, toBrokerData)
				if err != nil {
					logrus.Error("An error of transmitting data to the broker in data occurs: ", err)
				}
			}()
		}
	}
}
