package telebot

import (
	"context"
	"encoding/json"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	brokerEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/broker"
)

func processFromBrokerData(ctx context.Context,
	fromBrokerData brokerEnt.DataFrom,
	botConf BotConf) error {
	if fromBrokerData.IsCommand {
		var botNewComm botEnt.Command
		err := json.Unmarshal([]byte(fromBrokerData.Value), &botNewComm)
		if err != nil {
			return fmt.Errorf("failed to unmarshal a command object: %w", err)
		}

		return registerBotCommand(ctx, botNewComm, botConf)
	}

	toBotData, err := castFromBrokerData(fromBrokerData)
	if err != nil {
		return fmt.Errorf("failed to cast data from the broker: %w", err)
	}

	err = botConf.BotWorker.SendMessage(ctx, toBotData.Value, toBotData.ChatID)
	if err != nil {
		return fmt.Errorf("failed to send a message to the bot: %w", err)
	}

	return nil
}

func castFromBrokerData(data brokerEnt.DataFrom) (botEnt.Data, error) {

	var botData BotDataDto
	err := json.Unmarshal([]byte(data.Value), &botData)
	if err != nil {
		return botEnt.Data{}, fmt.Errorf("failed to unmarshal bot data: %w", err)
	}

	return botEnt.Data{ChatID: botData.ChatID, Value: botData.Value}, nil
}
