package telebot

import (
	"context"
	"encoding/json"
	"fmt"

	botEnt "github.com/Guise322/TeleBot/server/internal/entities/bot"
	serviceEnt "github.com/Guise322/TeleBot/server/internal/entities/service"
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

	brokerOutDataChan := transmitter.StartTransmittingData(ctx)
	botInDataChan := bot.Start(ctx)

	for {
		select {
		case <-ctx.Done():
			bot.Stop()
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

			brokerOutDataChan <- brokerOutData
		}
	}
}

func processBrokerInData(ctx context.Context,
	brokerInData serviceEnt.InData,
	bot botInterf.Worker,
	botCommands *[]botEnt.Command) {
	if brokerInData.IsCommand {
		var botNewComm botEnt.Command
		err := json.Unmarshal([]byte(brokerInData.Value), &botNewComm)
		if err != nil {
			logrus.Error("Can not unmarshal a command object: ", err)

			return
		}

		*botCommands = append(*botCommands, botNewComm)
		bot.RegisterCommands(ctx, *botCommands)
	}

	botOutData, err := castBrokerInData(brokerInData)
	if err != nil {
		logrus.Error("An error of casting broker in data occurs: ", err)

		return
	}

	bot.SendMessage(ctx, botOutData.Value, botOutData.ChatID)
}

func castBrokerInData(data serviceEnt.InData) (botEnt.Data, error) {

	var botData BotDataDto
	err := json.Unmarshal([]byte(data.Value), &botData)
	if err != nil {
		return botEnt.Data{}, fmt.Errorf("can not unmarshal a bot data: %w", err)
	}

	return botEnt.Data{ChatID: botData.ChatID, Value: botData.Value}, nil
}

func processBotInData(data botEnt.Data, commands []botEnt.Command) (serviceEnt.OutData, error) {
	if !data.IsCommand {
		chatID := data.ChatID
		message := data.Value
		dataDto := BotDataDto{ChatID: chatID, Value: message}
		dataValue, err := json.Marshal(dataDto)
		if err != nil {
			return serviceEnt.OutData{}, fmt.Errorf("can not marshal a BotDataDto with data: %w", err)
		}

		return serviceEnt.OutData{Value: string(dataValue)}, nil
	}

	for _, command := range commands {
		if data.Value != command.Name {
			continue
		}

		chatID := data.ChatID
		dataDto := BotDataDto{ChatID: chatID}
		dataValue, err := json.Marshal(dataDto)
		if err != nil {
			logrus.Error()

			return serviceEnt.OutData{}, fmt.Errorf("can not marshal a BotDataDto with a command: %w", err)
		}

		return serviceEnt.OutData{CommName: command.Name, Value: string(dataValue)}, nil
	}

	return serviceEnt.OutData{}, fmt.Errorf("no commands with the name %v", data.Value)
}
