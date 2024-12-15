package storage

import "github.com/DOs0x12/TeleBot/server/internal/entities/bot"

type CommandStorage interface {
	Save(comm bot.Command) error
	Load() ([]bot.Command, error)
	Close()
}
