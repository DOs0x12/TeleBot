package telebot

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
			botOutData := processBrokerInData(brokerInData, bot, botCommands)
			var zeroBotData botEnt.Data
			if botOutData != zeroBotData {
				go sendMessageWithRetries(ctx, bot, botOutData)
			}
		case botInData := <-botInDataChan:
			brokerOutData := processBotInData(botInData, *botCommands)
			var zeroBotOutData serviceEnt.OutData
			if brokerOutData != zeroBotOutData {
				brokerOutDataChan <- brokerOutData
			}
		}
	}
}

func processBrokerInData(data serviceEnt.InData,
	bot botInterf.Worker,
	botCommands *[]botEnt.Command) botEnt.Data {
	if data.IsCommand {
		var botNewComm botEnt.Command
		err := json.Unmarshal([]byte(data.Value), &botNewComm)
		if err != nil {
			logrus.Error("Can not unmarshal a command object:", err)

			return botEnt.Data{}
		}

		*botCommands = append(*botCommands, botNewComm)
		bot.RegisterCommands(*botCommands)

		return botEnt.Data{}
	}

	var botData BotDataDto
	err := json.Unmarshal([]byte(data.Value), &botData)
	if err != nil {
		logrus.Error("Can not unmarshal a bot data:", err)

		return botEnt.Data{}
	}

	return botEnt.Data{ChatID: botData.ChatID, Value: botData.Value}
}

func sendMessageWithRetries(ctx context.Context, bot botInterf.Worker, botOutData botEnt.Data) {
	numOfRetries := 10
	timeBetweenRetries := 5 * time.Second

	for i := 0; i < numOfRetries; i++ {
		if ctx.Err() != nil {
			return
		}

		err := bot.SendMessage(botOutData.Value, botOutData.ChatID)
		if err == nil {
			break
		}

		logrus.Error("Cannot send a message:", err)

		time.Sleep(timeBetweenRetries)
	}
}

func processBotInData(data botEnt.Data,
	commands []botEnt.Command) serviceEnt.OutData {

	if !data.IsCommand {
		chatID := data.ChatID
		message := data.Value
		dataDto := BotDataDto{ChatID: chatID, Value: message}
		dataValue, err := json.Marshal(dataDto)
		if err != nil {
			logrus.Error("Cannot marshal a BotDataDto with data:", err)

			return serviceEnt.OutData{}
		}

		return serviceEnt.OutData{Value: string(dataValue)}
	}

	for _, command := range commands {
		if data.Value != command.Name {
			continue
		}

		chatID := data.ChatID
		dataDto := BotDataDto{ChatID: chatID}
		dataValue, err := json.Marshal(dataDto)
		if err != nil {
			logrus.Error("Cannot marshal a BotDataDto with a command:", err)

			return serviceEnt.OutData{}
		}

		return serviceEnt.OutData{CommName: command.Name, Value: string(dataValue)}
	}

	return serviceEnt.OutData{}
}
