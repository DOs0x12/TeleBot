package bot

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/internal/common"
	botEnt "github.com/DOs0x12/TeleBot/server/internal/entities/bot"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telebot struct {
	bot      *tgbot.BotAPI
	commands *[]botEnt.Command
}

func NewTelebot(ctx context.Context, botKey string, commands []botEnt.Command) (Telebot, error) {
	botApi, err := tgbot.NewBotAPI(botKey)
	if err != nil {
		return Telebot{}, fmt.Errorf("an error of connecting to the bot occurs: %w", err)
	}

	bot := Telebot{bot: botApi, commands: &commands}
	err = bot.RegisterCommands(ctx, commands)
	if err != nil {
		return Telebot{}, fmt.Errorf("can not register commands in the bot: %w", err)
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

func (t Telebot) RegisterCommands(ctx context.Context, commands []botEnt.Command) error {
	conf := configureCommands(commands)

	resp, err := t.requestWithRetries(ctx, conf)
	if err != nil {
		return fmt.Errorf("an error of sending a request to Telegram occurs: %w with the result: %v", err, resp.Description)
	}

	return nil
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

func (t Telebot) SendMessage(ctx context.Context, msg string, chatID int64) error {
	tgMsg := tgbot.NewMessage(chatID, msg)
	err := t.sendWithRetries(ctx, tgMsg)
	if err != nil {
		return fmt.Errorf("can not send a message to the telegram bot: %w", err)
	}

	return nil
}

func (t Telebot) sendWithRetries(ctx context.Context, msg tgbot.MessageConfig) error {
	act := func(ctx context.Context) error {
		_, err := t.bot.Send(msg)

		return err
	}

	return common.ExecuteWithRetries(ctx, act)
}

func (t Telebot) requestWithRetries(ctx context.Context, conf tgbot.SetMyCommandsConfig) (*tgbot.APIResponse, error) {
	var resp *tgbot.APIResponse
	act := func(ctx context.Context) error {
		var err error
		resp, err = t.bot.Request(conf)

		return err
	}

	err := common.ExecuteWithRetries(ctx, act)

	return resp, err
}
