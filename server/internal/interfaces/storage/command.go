package storage

import (
	"context"

	"github.com/DOs0x12/TeleBot/server/v3/internal/entities/bot"
)

type CommandStorage interface {
	Save(ctx context.Context, comm bot.Command) error
	Load(ctx context.Context) ([]bot.Command, error)
	Close()
}
