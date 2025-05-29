package bot

import (
	"context"
	"fmt"
	"time"

	botEnt "github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
	"github.com/DOs0x12/TeleBot/server/v2/retry"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (t telebot) RegisterCommands(ctx context.Context, commands []botEnt.Command) error {
	conf := configureCommands(commands)

	resp, err := t.requestWithRetries(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to send a request to Telegram: %w with the result: %v", err, resp.Description)
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

func (t telebot) requestWithRetries(ctx context.Context, conf tgbot.SetMyCommandsConfig) (*tgbot.APIResponse, error) {
	var resp *tgbot.APIResponse
	act := func(ctx context.Context) error {
		var err error
		resp, err = t.bot.Request(conf)

		return err
	}
	rCnt := 5
	rDur := 1 * time.Second
	err := retry.ExecuteWithRetries(ctx, act, rCnt, rDur)

	return resp, err
}
