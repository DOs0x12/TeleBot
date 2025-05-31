package bot

import (
	"context"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v3/internal/entities/bot"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type telebot struct {
	bot      *tgbot.BotAPI
	commands *[]botEnt.Command
}

func NewTelebot(ctx context.Context, botKey string, commands []botEnt.Command) (telebot, error) {
	botApi, err := tgbot.NewBotAPI(botKey)
	if err != nil {
		return telebot{}, fmt.Errorf("failed to connect to the bot: %w", err)
	}

	bot := telebot{bot: botApi, commands: &commands}

	return bot, nil
}

func (t telebot) Start(ctx context.Context) (<-chan botEnt.Data, <-chan error) {
	updConfig := tgbot.NewUpdate(0)
	botInDataChan := make(chan botEnt.Data)
	botErrChan := make(chan error)
	updChan := t.bot.GetUpdatesChan(updConfig)

	go t.receiveInData(ctx, updChan, botInDataChan, botErrChan)

	return botInDataChan, botErrChan
}

func (t telebot) Stop() {
	t.bot.StopReceivingUpdates()
}
