package bot

import (
	"context"

	"github.com/Guise322/TeleBot/server/internal/entities/bot"
)

type Worker interface {
	Start(ctx context.Context) chan bot.Data
	Stop()
	SendMessage(msg string, chatID int64) error
	RegisterCommands(commands *[]bot.Command) error
}
