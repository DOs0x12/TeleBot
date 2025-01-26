package telebot

import (
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	"github.com/DOs0x12/TeleBot/server/v2/internal/entities/broker"
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
	searchBotCommandByName(botNewComm.Name, *s.botConf.BotCommands)
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
	commands []botEnt.Command) *botEnt.Command {
	for _, command := range commands {
		if commName == command.Name {
			return &command
		}
	}

	return nil
}

func (s service) processBotCommand(fromBrokerComm broker.CommandFrom) error {
	comm := botEnt.Command{
		Name:        fromBrokerComm.Name,
		Description: fromBrokerComm.Description,
		Token:       fromBrokerComm.Token,
	}
	err := s.registerBotCommand(comm)
	if err != nil {
		return fmt.Errorf("failed to register a command in the bot: %w", err)
	}

	return nil
}

func (s service) processBotData(fromBrokerData broker.DataFrom) error {
	err := s.botConf.BotWorker.SendMessage(s.ctx, fromBrokerData.Value, fromBrokerData.ChatID)
	if err != nil {
		return fmt.Errorf("failed to send a message to the bot: %w", err)
	}

	return nil
}
