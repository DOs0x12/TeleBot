package command

import (
	"github.com/Guise322/TeleBot/server/internal/entities/bot"
)

func CreateCommand(name, desc string) bot.Command {
	return bot.Command{
		Name:        name,
		Description: desc,
	}
}
