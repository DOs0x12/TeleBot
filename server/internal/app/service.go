package telebot

import (
	"context"
	"fmt"

	botEnt "github.com/Guise322/TeleBot/server/internal/entities/bot"
	botInterf "github.com/Guise322/TeleBot/server/internal/interfaces/bot"
	serviceInterf "github.com/Guise322/TeleBot/server/internal/interfaces/broker"

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
	receiver serviceInterf.DataReceiver,
	transmitter serviceInterf.DataTransmitter,
	botCommands *[]botEnt.Command) error {
	brokerInDataChan, err := receiver.StartReceivingData(ctx)
	if err != nil {
		return fmt.Errorf("an error of the data receiver occurs: %w", err)
	}

	botInDataChan := bot.Start(ctx)

	for {
		select {
		case <-ctx.Done():
			bot.Stop()
			transmitter.Close()
			logrus.Info("The bot is stopped")

			return nil
		case brokerInData := <-brokerInDataChan:
			go processBrokerInData(ctx, brokerInData, bot, botCommands)
		case botInData := <-botInDataChan:
			brokerOutData, err := processBotInData(botInData, *botCommands)
			if err != nil {
				logrus.Error("An error of processing bot in data occurs: ", err)

				continue
			}

			err = transmitter.TransmitData(ctx, brokerOutData)
			if err != nil {
				logrus.Error("An error of transmitting data to the broker in data occurs: ", err)

				continue
			}
		}
	}
}
