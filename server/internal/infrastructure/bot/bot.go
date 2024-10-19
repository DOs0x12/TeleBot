package bot

import (
	"context"
	"fmt"
	"time"

	botEnt "github.com/Guise322/TeleBot/server/internal/entities/bot"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telebot struct {
	bot      *tgbot.BotAPI
	commands *[]botEnt.Command
}

func NewTelebot(botKey string, commands *[]botEnt.Command) (Telebot, error) {
	botApi, err := tgbot.NewBotAPI(botKey)
	if err != nil {
		var zero Telebot

		return zero, fmt.Errorf("getting an error at connecting to the bot: %v", err)
	}

	botApi.Request(configureCommands(commands))

	return Telebot{bot: botApi, commands: commands}, nil
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

			botInDataChan <- botEnt.Data{ChatID: upd.Message.Chat.ID, Value: upd.Message.Text}
		}
	}
}

func (t Telebot) RegisterCommands(commands *[]botEnt.Command) error {
	conf := configureCommands(commands)

	if _, err := t.bot.Request(conf); err != nil {
		return fmt.Errorf("getting an error at registering commands: %v", err)
	}

	return nil
}

func configureCommands(commands *[]botEnt.Command) tgbot.SetMyCommandsConfig {
	commandSet := make([]tgbot.BotCommand, len(*commands))

	for i, command := range *commands {
		commandSet[i] = tgbot.BotCommand{Command: command.Name, Description: command.Description}
	}

	return tgbot.NewSetMyCommands(commandSet...)
}

func (t Telebot) Stop() {
	t.bot.StopReceivingUpdates()
}

func (t Telebot) SendMessage(msg string, chatID int64) error {
	tgMsg := tgbot.NewMessage(chatID, msg)

	maxRetries := 10
	cnt := 0
	var err error

	for cnt < maxRetries {
		if _, err = t.bot.Send(tgMsg); err != nil {
			time.Sleep(5 * time.Second)
			cnt++

			continue
		}

		return nil
	}

	return err
}
