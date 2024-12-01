package telebot

import (
	"context"
	"encoding/json"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/internal/entities/bot"
	brokerEnt "github.com/DOs0x12/TeleBot/server/internal/entities/broker"
	botInterf "github.com/DOs0x12/TeleBot/server/internal/interfaces/bot"
)

func processBrokerInData(ctx context.Context,
	brokerInData brokerEnt.InData,
	bot botInterf.Worker,
	botCommands *[]botEnt.Command) error {
	if brokerInData.IsCommand {
		var botNewComm botEnt.Command
		err := json.Unmarshal([]byte(brokerInData.Value), &botNewComm)
		if err != nil {
			return fmt.Errorf("can not unmarshal a command object: %w", err)
		}

		*botCommands = append(*botCommands, botNewComm)
		bot.RegisterCommands(ctx, *botCommands)
	}

	botOutData, err := castBrokerInData(brokerInData)
	if err != nil {
		return fmt.Errorf("an error of casting broker in data occurs: %w", err)
	}

	err = bot.SendMessage(ctx, botOutData.Value, botOutData.ChatID)
	if err != nil {
		return fmt.Errorf("an error of sending a message to the bot occurs: %w", err)
	}

	return nil
}

func castBrokerInData(data brokerEnt.InData) (botEnt.Data, error) {

	var botData BotDataDto
	err := json.Unmarshal([]byte(data.Value), &botData)
	if err != nil {
		return botEnt.Data{}, fmt.Errorf("can not unmarshal bot data: %w", err)
	}

	return botEnt.Data{ChatID: botData.ChatID, Value: botData.Value}, nil
}
