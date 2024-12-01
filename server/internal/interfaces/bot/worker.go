package bot

import (
	"context"

	"github.com/DOs0x12/TeleBot/server/internal/entities/bot"
)

type Worker interface {
	Start(ctx context.Context) chan bot.Data
	Stop()
	SendMessage(ctx context.Context, msg string, chatID int64) error
	RegisterCommands(ctx context.Context, commands []bot.Command) error
}
