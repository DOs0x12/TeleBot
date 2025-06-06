package telebot

import (
	"context"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v3/internal/entities/bot"
	"github.com/DOs0x12/TeleBot/server/v3/internal/entities/broker"
	botInterf "github.com/DOs0x12/TeleBot/server/v3/internal/interfaces/bot"
	brokerInterf "github.com/DOs0x12/TeleBot/server/v3/internal/interfaces/broker"
	storInterf "github.com/DOs0x12/TeleBot/server/v3/internal/interfaces/storage"

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

func NewService(
	ctx context.Context,
	botConf BotConf,
	msgBroker brokerInterf.MessageBroker,
) service {

	return service{ctx: ctx, botConf: botConf, msgBroker: msgBroker}
}

func (s service) Process() error {
	fromBrokerDataChan, fromBrokerCommChan, errChan := s.msgBroker.StartReceivingData(s.ctx)
	fromBotDataChan, botErrChan := s.botConf.BotWorker.Start(s.ctx)
	err := s.loadBotCommands()
	if err != nil {
		return err
	}

	err = s.botConf.BotWorker.RegisterCommands(s.ctx, *s.botConf.BotCommands)
	if err != nil {
		return fmt.Errorf("failed to register the loaded commands: %w", err)
	}

	s.handleServices(fromBrokerDataChan, fromBrokerCommChan, fromBotDataChan, errChan, botErrChan)
	s.stopServices()

	return nil
}

func (s service) handleServices(
	fromBrokerDataChan <-chan broker.DataFrom,
	fromBrokerCommChan <-chan broker.CommandFrom,
	fromBotDataChan <-chan botEnt.Data,
	receiverErrChan <-chan error,
	botErrChan <-chan error,
) {
	for {
		select {
		case fromBrokerData, ok := <-fromBrokerDataChan:
			if !ok {
				return
			}
			go s.processFromBrokerData(fromBrokerData)
		case fromBrokerComm, ok := <-fromBrokerCommChan:
			if !ok {
				return
			}
			go s.processFromBrokerCommand(fromBrokerComm)
		case recErr, ok := <-receiverErrChan:
			if !ok {
				return
			}
			logrus.Error("Failed to receive broker data: ", recErr)
		case fromBotData, ok := <-fromBotDataChan:
			if !ok {
				return
			}
			go s.processFromBotData(fromBotData)
		case botErr, ok := <-botErrChan:
			if !ok {
				return
			}
			logrus.Error("Failed to receive bot data: ", botErr)
		}
	}
}

func (s service) stopServices() {
	s.botConf.BotWorker.Stop()
	logrus.Info("The bot has been stopped")
	err := s.msgBroker.Close()
	if err != nil {
		logrus.Errorf("Failed to close the broker: %v", err)
	} else {
		logrus.Info("The transmitter connection has been closed")
	}
	s.botConf.Storage.Close()
	logrus.Info("The storage connection has been closed")
}

func (s service) processFromBrokerData(fromBrokerData broker.DataFrom) {
	err := s.processBotData(fromBrokerData)
	if err != nil {
		logrus.Error("Failed to process bot data: ", err)
	}

	err = s.msgBroker.Commit(s.ctx, fromBrokerData.MsgUuid)
	if err != nil {
		logrus.WithField("UUID", fromBrokerData.MsgUuid).
			Error("failed to commit a message: ", err)
	}
}

func (s service) processFromBrokerCommand(fromBrokerComm broker.CommandFrom) {
	err := s.processBotCommand(fromBrokerComm)
	if err != nil {
		logrus.Error("Failed to process a bot command: ", err)
	}

	err = s.msgBroker.Commit(s.ctx, fromBrokerComm.MsgUuid)
	if err != nil {
		logrus.WithField("UUID", fromBrokerComm.MsgUuid).
			Error("failed to commit a message with a command: ", err)
	}
}

func (s service) toBrokerData(fromBotData botEnt.Data) (broker.DataTo, error) {
	if !fromBotData.IsCommand {
		return broker.DataTo{
			ChatID: fromBotData.ChatID,
			Value:  fromBotData.Value,
			IsFile: fromBotData.IsFile,
		}, nil
	}

	botCommand := searchBotCommandByName(string(fromBotData.Value), s.botConf.BotCommands)
	if botCommand == nil {
		return broker.DataTo{}, fmt.Errorf("no commands with the name %v", fromBotData.Value)
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
