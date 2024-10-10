package command

import (
	"TeleBot/internal/entities/bot"
)

func CreateHelloCommand() bot.Command {
	return bot.Command{
		Name:        "/hello1",
		Description: "Say hello to the bot",
	}
}

func CreateHelloMonsiuerCommand() bot.Command {
	return bot.Command{
		Name:        "/hello2",
		Description: "Say hello to the monsiuer",
	}
}

func CreateCommand(name, desc string) bot.Command {
	return bot.Command{
		Name:        name,
		Description: desc,
	}
}
