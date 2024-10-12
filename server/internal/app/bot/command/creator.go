package command

import (
	"TeleBot/internal/entities/bot"
)

func CreateCommand(name, desc string) bot.Command {
	return bot.Command{
		Name:        name,
		Description: desc,
	}
}
