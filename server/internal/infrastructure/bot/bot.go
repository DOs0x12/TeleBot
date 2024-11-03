package bot

import (
	"context"
	"fmt"

	botEnt "github.com/Guise322/TeleBot/server/internal/entities/bot"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telebot struct {
	bot      *tgbot.BotAPI
	commands *[]botEnt.Command
}

func NewTelebot(botKey string, commands []botEnt.Command) (Telebot, error) {
	botApi, err := tgbot.NewBotAPI(botKey)
	if err != nil {
		return Telebot{}, fmt.Errorf("an error of connecting to the bot occurs: %w", err)
	}

	resp, err := botApi.Request(configureCommands(commands))
	if err != nil {
		return Telebot{}, fmt.Errorf("register commands of a new bot: get an response from the bot API %v wiht an error: %w",
			resp.Result, err)
	}

	return Telebot{bot: botApi, commands: &commands}, nil
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

func (t Telebot) RegisterCommands(commands []botEnt.Command) error {
	conf := configureCommands(commands)

	resp, err := t.bot.Request(conf)
	if err != nil {
		err = fmt.Errorf("register commands: get an response from the bot API %v wiht an error: %w",
			resp.Result, err)
	}

	return err
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

func (t Telebot) SendMessage(msg string, chatID int64) error {
	tgMsg := tgbot.NewMessage(chatID, msg)

	_, err := t.bot.Send(tgMsg)
	if err != nil {
		err = fmt.Errorf("an error of sending a message occurs: %w", err)
	}

	return err
}
