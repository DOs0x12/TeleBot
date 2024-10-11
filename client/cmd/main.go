package main

import (
	"Telebot/client/pkg/service"
	"context"
	"os/signal"
	"strings"
	"syscall"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	kafkaAddr := "localhost:9092"
	w := &kafka.Writer{
		Addr:     kafka.TCP(kafkaAddr),
		Balancer: &kafka.LeastBytes{},
	}

	commData := service.CommandData{Name: "/test", Description: "testCommand"}

	err := service.RegisterCommand(ctx, w, commData)
	if err != nil {
		logrus.Error("Failed to register a command:", err)

		return
	}

	readingTopic := strings.Trim(commData.Name, "/")
	inDataChan := service.StartGetData(ctx, readingTopic, kafkaAddr)

	for {
		select {
		case <-ctx.Done():
			return
		case inData := <-inDataChan:
			outData := service.BotData{ChatID: inData.ChatID, Value: "Test!"}
			service.SendData(ctx, w, outData)
		}
	}
}
