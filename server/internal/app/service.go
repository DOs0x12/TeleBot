package telebot

import (
	"context"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	"github.com/DOs0x12/TeleBot/server/v2/internal/entities/broker"
	botInterf "github.com/DOs0x12/TeleBot/server/v2/internal/interfaces/bot"
	brokerInterf "github.com/DOs0x12/TeleBot/server/v2/internal/interfaces/broker"
	storInterf "github.com/DOs0x12/TeleBot/server/v2/internal/interfaces/storage"

	"github.com/sirupsen/logrus"
)

type BotConf struct {
	BotWorker   botInterf.Worker
	BotCommands *[]botEnt.Command
	Storage     storInterf.CommandStorage
}

type service struct {
	ctx       context.Context
	botConf   BotConf
	msgBroker brokerInterf.MessageBroker
}

func NewService(ctx context.Context,
	botConf BotConf,
	msgBroker brokerInterf.MessageBroker) service {

	return service{ctx: ctx, botConf: botConf, msgBroker: msgBroker}
}

func (s service) Process() error {
	fromBrokerDataChan, err := s.msgBroker.StartReceivingData(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to start receiving data: %w", err)
	}

	fromBotDataChan := s.botConf.BotWorker.Start(s.ctx)
	if err = s.loadBotCommands(); err != nil {
		return err
	}

	if err = s.botConf.BotWorker.RegisterCommands(s.ctx, *s.botConf.BotCommands); err != nil {
		return fmt.Errorf("failed to register the loaded commands: %w", err)
	}

	s.handleServices(fromBrokerDataChan, fromBotDataChan)

	return nil
}

func (s service) handleServices(fromBrokerDataChan <-chan broker.DataFrom,
	fromBotDataChan <-chan botEnt.Data) {
	for {
		select {
		case <-s.ctx.Done():
			s.stopServices()

			return
		case fromBrokerData := <-fromBrokerDataChan:
			go s.processFromBrokerData(fromBrokerData)
		case fromBotData := <-fromBotDataChan:
			go s.processFromBotData(fromBotData)
		}
	}
}

func (s service) stopServices() {
	s.botConf.BotWorker.Stop()
	logrus.Info("The bot has been stopped")
	s.msgBroker.Close()
	logrus.Info("The transmitter connection has been closed")
	s.botConf.Storage.Close()
	logrus.Info("The storage connection has been closed")
}

func (s service) processFromBrokerData(fromBrokerData broker.DataFrom) {
	if fromBrokerData.IsCommand {
		err := s.processBotCommand(fromBrokerData)
		if err != nil {
			logrus.Error("Failed to process a bot command: ", err)
		}

		return
	}

	err := s.processBotData(fromBrokerData)
	if err != nil {
		logrus.Error("Failed to process bot data: ", err)
	}
}

func (s service) processBotCommand(fromBrokerData broker.DataFrom) error {
	comm, err := unmarshalBotCommand(fromBrokerData.Value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal a bot command: %w", err)
	}

	err = s.registerBotCommand(comm)
	if err != nil {
		return fmt.Errorf("failed to register a command in the bot: %w", err)
	}

	return nil
}

func (s service) processBotData(fromBrokerData broker.DataFrom) error {
	toBotData, err := unmarshalBotData(fromBrokerData.Value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bot data: %w", err)
	}

	err = s.botConf.BotWorker.SendMessage(s.ctx, toBotData.Value, toBotData.ChatID)
	if err != nil {
		return fmt.Errorf("failed to send a message to the bot: %w", err)
	}

	err = s.msgBroker.Commit(s.ctx, fromBrokerData.MsgUuid)
	if err != nil {
		return fmt.Errorf("failed to commit the message with UUID %v: %w", fromBrokerData.MsgUuid, err)
	}

	return nil
}

func (s service) toBrokerData(fromBotData botEnt.Data) (broker.DataTo, error) {
	if !fromBotData.IsCommand {
		return broker.DataTo{ChatID: fromBotData.ChatID, Value: fromBotData.Value}, nil
	}

	botCommand, err := searchBotCommandByName(fromBotData.Value, *s.botConf.BotCommands)
	if err != nil {
		return broker.DataTo{}, err
	}

	return broker.DataTo{
			CommName: botCommand.Name,
			ChatID:   fromBotData.ChatID,
			Value:    fromBotData.Value,
			Token:    botCommand.Token,
		},
		nil
}

func (s service) processFromBotData(fromBotData botEnt.Data) {
	toBrokerData, err := s.toBrokerData(fromBotData)
	if err != nil {
		logrus.Error("Failed to get a bot command: ", err)

		return
	}

	err = s.msgBroker.TransmitData(s.ctx, toBrokerData)
	if err != nil {
		logrus.Error("Failed to transmit data to the broker: ", err)
	}
}
