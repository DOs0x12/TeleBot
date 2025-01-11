package bot

import (
	"context"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telebot struct {
	bot      *tgbot.BotAPI
	commands *[]botEnt.Command
}

func NewTelebot(ctx context.Context, botKey string, commands []botEnt.Command) (Telebot, error) {
	botApi, err := tgbot.NewBotAPI(botKey)
	if err != nil {
		return Telebot{}, fmt.Errorf("failed to connect to the bot: %w", err)
	}

	bot := Telebot{bot: botApi, commands: &commands}
	err = bot.RegisterCommands(ctx, commands)
	if err != nil {
		return Telebot{}, fmt.Errorf("failed to register commands in the bot: %w", err)
	}

	return bot, nil
}

func (t Telebot) Start(ctx context.Context) chan botEnt.Data {
	updConfig := tgbot.NewUpdate(0)
	botInDataChan := make(chan botEnt.Data)
	updChan := t.bot.GetUpdatesChan(updConfig)

	go receiveInData(ctx, updChan, botInDataChan)

	return botInDataChan
}

func (t Telebot) Stop() {
	t.bot.StopReceivingUpdates()
}
