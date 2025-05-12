package bot

import (
	"context"

	"github.com/DOs0x12/TeleBot/server/v2/internal/entities/bot"
)

type Worker interface {
	Start(ctx context.Context) (<-chan bot.Data, <-chan error)
	Stop()
	SendMessage(ctx context.Context, msg string, chatID int64) error
	SendDocument(ctx context.Context, chatID int64, docData, docName string) error
	RegisterCommands(ctx context.Context, commands []bot.Command) error
}
