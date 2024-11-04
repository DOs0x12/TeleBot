package telebot

import (
	"context"
	"encoding/json"
	"fmt"

	botEnt "github.com/Guise322/TeleBot/server/internal/entities/bot"
	serviceEnt "github.com/Guise322/TeleBot/server/internal/entities/service"
	botInterf "github.com/Guise322/TeleBot/server/internal/interfaces/bot"
	"github.com/sirupsen/logrus"
)

func processBrokerInData(ctx context.Context,
	brokerInData serviceEnt.InData,
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

	bot.SendMessage(ctx, botOutData.Value, botOutData.ChatID)
}

func castBrokerInData(data serviceEnt.InData) (botEnt.Data, error) {

	var botData BotDataDto
	err := json.Unmarshal([]byte(data.Value), &botData)
	if err != nil {
		return botEnt.Data{}, fmt.Errorf("can not unmarshal a bot data: %w", err)
	}

	return botEnt.Data{ChatID: botData.ChatID, Value: botData.Value}, nil
}
