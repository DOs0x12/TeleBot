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
		return fmt.Errorf("failed to load the bot commands from the storage: %w", err)
	}

	*botCommands = append(*botCommands, commands...)

	return nil
}

func registerBotCommand(ctx context.Context,
	botNewComm botEnt.Command,
	botConf BotConf) error {
	*botConf.BotCommands = append(*botConf.BotCommands, botNewComm)
	botConf.BotWorker.RegisterCommands(ctx, *botConf.BotCommands)

	err := botConf.Storage.Save(ctx, botNewComm)
	if err != nil {
		return fmt.Errorf("failed to save a command: %w", err)
	}

	return nil
}

func processFromBotData(
	data botEnt.Data,
	commands []botEnt.Command) (brokerEnt.DataTo, error) {
	if !data.IsCommand {
		return processFromBotCommand(data)
	}

	for _, command := range commands {
		if data.Value != command.Name {
			continue
		}

		chatID := data.ChatID
		dataDto := BotDataDto{ChatID: chatID}
		dataValue, err := json.Marshal(dataDto)
		if err != nil {
			return brokerEnt.DataTo{}, fmt.Errorf("failed to marshal a BotDataDto with a command: %w", err)
		}

		return brokerEnt.DataTo{CommName: command.Name, Value: string(dataValue)}, nil
	}

	return brokerEnt.DataTo{}, fmt.Errorf("no commands with the name %v", data.Value)
}

func processFromBotCommand(data botEnt.Data) (brokerEnt.DataTo, error) {
	chatID := data.ChatID
	message := data.Value
	dataDto := BotDataDto{ChatID: chatID, Value: message}
	dataValue, err := json.Marshal(dataDto)
	if err != nil {
		return brokerEnt.DataTo{}, fmt.Errorf("failed to marshal a BotDataDto with data: %w", err)
	}

	return brokerEnt.DataTo{Value: string(dataValue)}, nil
}
