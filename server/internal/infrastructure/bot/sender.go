package bot

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/internal/common/retry"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

	return retry.ExecuteWithRetries(ctx, act)
}

func (t Telebot) requestWithRetries(ctx context.Context, conf tgbot.SetMyCommandsConfig) (*tgbot.APIResponse, error) {
	var resp *tgbot.APIResponse
	act := func(ctx context.Context) error {
		var err error
		resp, err = t.bot.Request(conf)

		return err
	}

	err := retry.ExecuteWithRetries(ctx, act)

	return resp, err
}