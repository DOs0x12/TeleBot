package telebot

import (
	"context"
	"encoding/json"
	"fmt"

	botEnt "github.com/Guise322/TeleBot/server/internal/entities/bot"
	brokerEnt "github.com/Guise322/TeleBot/server/internal/entities/broker"
	botInterf "github.com/Guise322/TeleBot/server/internal/interfaces/bot"
	"github.com/sirupsen/logrus"
)

func processBrokerInData(ctx context.Context,
	brokerInData brokerEnt.InData,
	bot botInterf.Worker,
	botCommands *[]botEnt.Command) {
	if brokerInData.IsCommand {
		var botNewComm botEnt.Command
		err := json.Unmarshal([]byte(brokerInData.Value), &botNewComm)
		if err != nil {
			logrus.Error("Can not unmarshal a command object: ", err)

			return
		}

		*botCommands = append(*botCommands, botNewComm)
		bot.RegisterCommands(ctx, *botCommands)
	}

	botOutData, err := castBrokerInData(brokerInData)
	if err != nil {
		logrus.Error("An error of casting broker in data occurs: ", err)

		return
	}

	err = bot.SendMessage(ctx, botOutData.Value, botOutData.ChatID)
	if err != nil {
		logrus.Error("An error of sending a message to the bot occurs: ", err)

		return
	}
}

func castBrokerInData(data brokerEnt.InData) (botEnt.Data, error) {

	var botData BotDataDto
	err := json.Unmarshal([]byte(data.Value), &botData)
	if err != nil {
		return botEnt.Data{}, fmt.Errorf("can not unmarshal bot data: %w", err)
	}

	return botEnt.Data{ChatID: botData.ChatID, Value: botData.Value}, nil
}
