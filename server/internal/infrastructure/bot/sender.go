package bot

import (
	"context"
	"fmt"

	"github.com/DOs0x12/TeleBot/server/v2/internal/common/retry"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (t telebot) SendMessage(ctx context.Context, msg string, chatID int64) error {
	tgMsg := tgbot.NewMessage(chatID, msg)
	err := t.sendWithRetries(ctx, tgMsg)
	if err != nil {
		return fmt.Errorf("failed to send a message to the telegram bot: %w", err)
	}

	return nil
}

func (t telebot) SendDocument(ctx context.Context, chatID int64, docData []byte, docName string) error {
	f := tgbot.FileBytes{Name: docName, Bytes: []byte(docData)}
	doc := tgbot.NewDocument(chatID, f)

	_, err := t.bot.Send(doc)
	if err != nil {
		return fmt.Errorf("failed to send a file %v: %w", docName, err)
	}

	return nil
}

func (t telebot) sendWithRetries(ctx context.Context, msg tgbot.MessageConfig) error {
	act := func(ctx context.Context) error {
		_, err := t.bot.Send(msg)

		return err
	}

	return retry.ExecuteWithRetries(ctx, act)
}
