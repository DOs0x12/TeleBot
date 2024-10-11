package main

import (
	"Telebot/client/pkg/service"
	"context"
	"os/signal"
	"syscall"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	w := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Balancer: &kafka.LeastBytes{},
	}

	commData := service.CommandData{Name: "/test", Description: "testCommand"}

	err := service.RegisterCommand(ctx, w, commData)
	if err != nil {
		logrus.Error("Failed to register a command:", err)

		return
	}
}
