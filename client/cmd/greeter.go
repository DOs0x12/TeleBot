package main

import (
	"context"
	"flag"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Guise322/TeleBot/client/pkg/service"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	kafkaAddr := flag.String("conn", "kafka:9092", "The kafka connection string.")
	flag.Parse()

	w := &kafka.Writer{
		Addr:     kafka.TCP(*kafkaAddr),
		Balancer: &kafka.LeastBytes{},
	}

	commData := service.CommandData{Name: "/hello", Description: "Say hello to the bot"}

	err := service.RegisterCommand(ctx, w, commData)
	if err != nil {
		logrus.Error("Failed to register a command:", err)

		return
	}

	readingTopic := strings.Trim(commData.Name, "/")
	inDataChan := service.StartGetData(ctx, readingTopic, *kafkaAddr)

	for {
		select {
		case <-ctx.Done():
			return
		case inData := <-inDataChan:
			outData := service.BotData{ChatID: inData.ChatID, Value: "Hello there!"}
			service.SendData(ctx, w, outData)
		}
	}
}
