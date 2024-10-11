package bot

import (
	"TeleBot/internal/entities/bot"
	"context"
)

type Worker interface {
	Start(ctx context.Context) chan bot.Data
	Stop()
	SendMessage(msg string, chatID int64) error
	RegisterCommands(commands *[]bot.Command) error
}
