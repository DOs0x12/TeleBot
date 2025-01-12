package telebot

import (
	"context"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	brokerEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/broker"
	"github.com/DOs0x12/TeleBot/server/v2/internal/interfaces/storage"
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
		return processFromBotMessage(data), nil
	}

	for _, command := range commands {
		if data.Value != command.Name {
			continue
		}

		return brokerEnt.DataTo{
			CommName: command.Name,
			ChatID:   data.ChatID,
			Value:    data.Value,
			Token:    command.Token,
		}, nil
	}

	return brokerEnt.DataTo{}, fmt.Errorf("no commands with the name %v", data.Value)
}

func processFromBotMessage(data botEnt.Data) brokerEnt.DataTo {
	return brokerEnt.DataTo{ChatID: data.ChatID, Value: data.Value}
}
