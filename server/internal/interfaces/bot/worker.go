package bot

import (
	"context"

	"github.com/Guise322/TeleBot/server/internal/entities/bot"
)

type Worker interface {
	Start(ctx context.Context) chan bot.Data
	Stop()
	SendMessage(ctx context.Context, msg string, chatID int64)
	RegisterCommands(ctx context.Context, commands []bot.Command) bool
}
