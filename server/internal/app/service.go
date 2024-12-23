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

func Process(ctx context.Context,
	bot botInterf.Worker,
	receiver brokerInterf.DataReceiver,
	transmitter brokerInterf.DataTransmitter,
	botCommands *[]botEnt.Command,
	storage storInterf.CommandStorage) error {
	brokerInDataChan, err := receiver.StartReceivingData(ctx)
	if err != nil {
		return fmt.Errorf("an error of the data receiver occurs: %w", err)
	}

	botInDataChan := bot.Start(ctx)
	if err = loadBotCommands(ctx, botCommands, storage); err != nil {
		return fmt.Errorf("can not load the bot commands: %w", err)
	}

	if err = bot.RegisterCommands(ctx, *botCommands); err != nil {
		return fmt.Errorf("can not register the loaded commands: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			bot.Stop()
			transmitter.Close()
			logrus.Info("The bot is stopped")

			return nil
		case brokerInData := <-brokerInDataChan:
			go func() {
				err := processBrokerInData(ctx, brokerInData, bot, botCommands, storage)
				if err != nil {
					logrus.Error("Can not process data from the broker: ", err)
				}

				err = receiver.Commit(ctx, brokerInData.MsgUuid)
				if err != nil {
					logrus.WithField("messageUuid", err).Error("Can not commit the message with UUID: ", err)
				}
			}()
		case botInData := <-botInDataChan:
			brokerOutData, err := processBotInData(botInData, *botCommands)
			if err != nil {
				logrus.Error("An error of processing bot in data occurs: ", err)

				continue
			}

			go func() {
				err = transmitter.TransmitData(ctx, brokerOutData)
				if err != nil {
					logrus.Error("An error of transmitting data to the broker in data occurs: ", err)
				}
			}()
		}
	}
}
