package bot

import (
	"context"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func receiveInData(ctx context.Context,
	updChan tgbot.UpdatesChannel,
	botInDataChan chan<- botEnt.Data) {
	for {
		select {
		case <-ctx.Done():
			return
		case upd := <-updChan:
			processBotData(botInDataChan, upd)
		}
	}
}

func processBotData(botInDataChan chan<- botEnt.Data, upd tgbot.Update) {
	if upd.Message == nil {
		return
	}

	isComm := upd.Message.Command() != ""

	botInDataChan <- botEnt.Data{
		ChatID:    upd.Message.Chat.ID,
		Value:     upd.Message.Text,
		IsCommand: isComm,
	}
}
