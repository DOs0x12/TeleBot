package telebot

import (
	"context"
	"encoding/json"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/internal/entities/bot"
	brokerEnt "github.com/DOs0x12/TeleBot/server/internal/entities/broker"
	"github.com/DOs0x12/TeleBot/server/internal/interfaces/storage"
)

func loadBotCommands(
	ctx context.Context,
	botCommands *[]botEnt.Command,
	storage storage.CommandStorage) error {
	commands, err := storage.Load(ctx)
	if err != nil {
		return fmt.Errorf("can not load ")
	}

	*botCommands = append(*botCommands, commands...)

	return nil
}

func processBotInData(
	ctx context.Context,
	data botEnt.Data,
	commands []botEnt.Command,
	storage storage.CommandStorage) (brokerEnt.OutData, error) {
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
			return brokerEnt.OutData{}, fmt.Errorf("can not marshal a BotDataDto with a command: %w", err)
		}

		newComm := command
		newComm.Description = string(dataValue)

		err = storage.Save(ctx, newComm)
		if err != nil {
			return brokerEnt.OutData{}, fmt.Errorf("can not save a command: %w", err)
		}

		return brokerEnt.OutData{CommName: newComm.Name, Value: newComm.Description}, nil
	}

	return brokerEnt.OutData{}, fmt.Errorf("no commands with the name %v", data.Value)
}
