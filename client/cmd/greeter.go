package main

import (
	"context"
	"flag"
	"os/signal"
	"strings"
	"syscall"

	service "github.com/Guise322/TeleBot/client/pkg/broker"

	"github.com/sirupsen/logrus"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	kafkaAddr := flag.String("conn", "kafka:9092", "The kafka connection string.")
	flag.Parse()

	b := service.NewBroker(*kafkaAddr)

	commData := service.CommandData{Name: "/hello", Description: "Say hello to the bot"}

	err := b.RegisterCommand(ctx, commData)
	if err != nil {
		logrus.Error("Failed to register a command: ", err)

		return
	}

	readingTopic := strings.Trim(commData.Name, "/")
	inDataChan := b.StartGetData(ctx, readingTopic, *kafkaAddr)

	for {
		select {
		case <-ctx.Done():
			return
		case inData := <-inDataChan:
			outData := service.BotData{ChatID: inData.ChatID, Value: "Hello there!"}
			b.SendData(ctx, outData)
		}
	}
}
