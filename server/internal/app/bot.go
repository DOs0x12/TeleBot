package telebot

import (
	"encoding/json"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/internal/entities/bot"
	brokerEnt "github.com/DOs0x12/TeleBot/server/internal/entities/broker"
	"github.com/sirupsen/logrus"
)

func processBotInData(data botEnt.Data, commands []botEnt.Command) (brokerEnt.OutData, error) {
	if !data.IsCommand {
		chatID := data.ChatID
		message := data.Value
		dataDto := BotDataDto{ChatID: chatID, Value: message}
		dataValue, err := json.Marshal(dataDto)
		if err != nil {
			return brokerEnt.OutData{}, fmt.Errorf("can not marshal a BotDataDto with data: %w", err)
		}

		return brokerEnt.OutData{Value: string(dataValue)}, nil
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

			return brokerEnt.OutData{}, fmt.Errorf("can not marshal a BotDataDto with a command: %w", err)
		}

		return brokerEnt.OutData{CommName: command.Name, Value: string(dataValue)}, nil
	}

	return brokerEnt.OutData{}, fmt.Errorf("no commands with the name %v", data.Value)
}
