package bot

import (
	"context"
	"fmt"

	botEnt "github.com/DOs0x12/TeleBot/server/internal/entities/bot"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
