package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Guise322/TeleBot/server/common"
	botEnt "github.com/Guise322/TeleBot/server/internal/entities/bot"
	"github.com/sirupsen/logrus"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telebot struct {
	bot      *tgbot.BotAPI
	commands *[]botEnt.Command
}

const numOfRetries = 10
const timeBetweenRetries = 5 * time.Second

func NewTelebot(ctx context.Context, botKey string, commands []botEnt.Command) (Telebot, error) {
	botApi, err := tgbot.NewBotAPI(botKey)
	if err != nil {
		return Telebot{}, fmt.Errorf("an error of connecting to the bot occurs: %w", err)
	}

	bot := Telebot{bot: botApi, commands: &commands}
	if !bot.RegisterCommands(ctx, commands) {
		return Telebot{}, errors.New("can not register commands in the bot")
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

func receiveInData(ctx context.Context,
	updChan tgbot.UpdatesChannel,
	botInDataChan chan botEnt.Data) {
	for {
		select {
		case <-ctx.Done():
			return
		case upd := <-updChan:
			if upd.Message == nil {
				continue
			}

			isComm := upd.Message.Command() != ""
			botInDataChan <- botEnt.Data{ChatID: upd.Message.Chat.ID,
				Value:     upd.Message.Text,
				IsCommand: isComm}
		}
	}
}

func (t Telebot) RegisterCommands(ctx context.Context, commands []botEnt.Command) bool {
	conf := configureCommands(commands)

	for i := 0; i < numOfRetries; i++ {
		if ctx.Err() != nil {
			return true
		}

		resp, err := t.bot.Request(conf)
		if err == nil {
			return true
		}

		logrus.WithField("bot_result", resp.Result).Error("An error occurs when register commands: ", err)

		common.WaitWithContext(ctx, timeBetweenRetries)
	}

	return false
}

func configureCommands(commands []botEnt.Command) tgbot.SetMyCommandsConfig {
	commandSet := make([]tgbot.BotCommand, len(commands))

	for i, command := range commands {
		commandSet[i] = tgbot.BotCommand{Command: command.Name, Description: command.Description}
	}

	return tgbot.NewSetMyCommands(commandSet...)
}

func (t Telebot) Stop() {
	t.bot.StopReceivingUpdates()
}

func (t Telebot) SendMessage(ctx context.Context, msg string, chatID int64) {
	for i := 0; i < numOfRetries; i++ {
		if ctx.Err() != nil {
			return
		}

		tgMsg := tgbot.NewMessage(chatID, msg)
		_, err := t.bot.Send(tgMsg)
		if err == nil {
			return
		}

		logrus.Error("Can not send a message to the bot: ", err)

		common.WaitWithContext(ctx, timeBetweenRetries)
	}
}
