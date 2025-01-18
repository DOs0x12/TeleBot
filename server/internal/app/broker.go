package telebot

import (
	"encoding/json"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	brokerEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/broker"
)

func castFromBrokerData(data brokerEnt.DataFrom) (botEnt.Data, error) {

	var botData BotDataDto
	err := json.Unmarshal([]byte(data.Value), &botData)
	if err != nil {
		return botEnt.Data{}, fmt.Errorf("failed to unmarshal bot data: %w", err)
	}

	return botEnt.Data{ChatID: botData.ChatID, Value: botData.Value}, nil
}
