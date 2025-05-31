package telebot

import (
	"encoding/json"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v3/internal/entities/bot"
	"github.com/DOs0x12/TeleBot/server/v3/internal/entities/broker"
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
	comm := searchBotCommandByName(botNewComm.Name, s.botConf.BotCommands)
	if comm == nil {
		*s.botConf.BotCommands = append(*s.botConf.BotCommands, botNewComm)
	} else {
		comm.Name = botNewComm.Name
		comm.Description = botNewComm.Description
		comm.Token = botNewComm.Token
	}

	s.botConf.BotWorker.RegisterCommands(s.ctx, *s.botConf.BotCommands)

	err := s.botConf.Storage.Save(s.ctx, botNewComm)
	if err != nil {
		return fmt.Errorf("failed to save a command: %w", err)
	}

	return nil
}

func searchBotCommandByName(
	commName string,
	commands *[]botEnt.Command,
) *botEnt.Command {
	for _, command := range *commands {
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

type FileDto struct {
	Name string
	Data []byte
}

func (s service) processBotData(fromBrokerData broker.DataFrom) error {
	if fromBrokerData.IsFile {
		var fd FileDto
		err := json.Unmarshal([]byte(fromBrokerData.Value), &fd)
		if err != nil {
			return fmt.Errorf("failed to unmarshal file data from broker: %w", err)
		}

		err = s.botConf.BotWorker.SendDocument(s.ctx, fromBrokerData.ChatID, fd.Data, fd.Name)
		if err != nil {
			return err
		}

		return nil
	}
	err := s.botConf.BotWorker.SendMessage(s.ctx, string(fromBrokerData.Value), fromBrokerData.ChatID)
	if err != nil {
		return fmt.Errorf("failed to send a message to the bot: %w", err)
	}

	return nil
}
