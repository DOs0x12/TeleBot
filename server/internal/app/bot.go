package telebot

import (
	"encoding/json"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
)

func (s service) loadBotCommands() error {
	commands, err := s.botConf.Storage.Load(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to load the bot commands from the storage: %w", err)
	}

	*s.botConf.BotCommands = append(*s.botConf.BotCommands, commands...)

	return nil
}

func (s service) registerBotCommand(botNewComm botEnt.Command) error {
	*s.botConf.BotCommands = append(*s.botConf.BotCommands, botNewComm)
	s.botConf.BotWorker.RegisterCommands(s.ctx, *s.botConf.BotCommands)

	err := s.botConf.Storage.Save(s.ctx, botNewComm)
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
