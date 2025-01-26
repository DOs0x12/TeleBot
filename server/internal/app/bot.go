package telebot

import (
	"context"
	"encoding/json"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
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

func searchBotCommandByName(
	commName string,
	commands []botEnt.Command) (botEnt.Command, error) {
	for _, command := range commands {
		if commName == command.Name {
			return command, nil
		}
	}

	return botEnt.Command{}, fmt.Errorf("no commands with the name %v", commName)
}

func unmarshalBotCommand(jsonComm string) (botEnt.Command, error) {
	var comm botEnt.Command
	err := json.Unmarshal([]byte(jsonComm), &comm)
	if err != nil {
		return botEnt.Command{}, fmt.Errorf("failed to unmarshal a command object: %w", err)
	}

	return comm, nil
}

type BotDataDto struct {
	ChatID int64
	Value  string
}

func unmarshalBotData(jsonBotData string) (botEnt.Data, error) {
	var botData BotDataDto
	err := json.Unmarshal([]byte(jsonBotData), &botData)
	if err != nil {
		return botEnt.Data{}, err
	}

	return botEnt.Data{ChatID: botData.ChatID, Value: botData.Value}, nil
}
