package telebot

import (
	"context"
	"encoding/json"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/internal/entities/bot"
	brokerEnt "github.com/DOs0x12/TeleBot/server/internal/entities/broker"
)

func processFromBrokerData(ctx context.Context,
	fromBrokerData brokerEnt.DataFrom,
	botConf BotConf) error {
	if fromBrokerData.IsCommand {
		var botNewComm botEnt.Command
		err := json.Unmarshal([]byte(fromBrokerData.Value), &botNewComm)
		if err != nil {
			return fmt.Errorf("can not unmarshal a command object: %w", err)
		}

		return registerBotCommand(ctx, botNewComm, botConf)
	}

	toBotData, err := castFromBrokerData(fromBrokerData)
	if err != nil {
		return fmt.Errorf("an error of casting data from broker occurs: %w", err)
	}

	err = botConf.BotWorker.SendMessage(ctx, toBotData.Value, toBotData.ChatID)
	if err != nil {
		return fmt.Errorf("an error of sending a message to the bot occurs: %w", err)
	}

	return nil
}

func castFromBrokerData(data brokerEnt.DataFrom) (botEnt.Data, error) {

	var botData BotDataDto
	err := json.Unmarshal([]byte(data.Value), &botData)
	if err != nil {
		return botEnt.Data{}, fmt.Errorf("can not unmarshal bot data: %w", err)
	}

	return botEnt.Data{ChatID: botData.ChatID, Value: botData.Value}, nil
}
