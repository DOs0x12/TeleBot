package main

import (
	"Telebot/client/pkg/service"
	"context"
	"encoding/json"
	"os/signal"
	"syscall"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type CommandData struct {
	Name        string
	Description string
}

func main() {
	ctx := context.Background()
	signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	w := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Balancer: &kafka.LeastBytes{},
	}

	commData := CommandData{Name: "/test", Description: "testCommand"}
	data, err := json.Marshal(commData)
	if err != nil {
		logrus.Error("Can not marshal command data object:", err)

		return
	}

	err = service.RegisterCommand(ctx, w, data)
	if err != nil {
		logrus.Error("Failed to register a command:", err)

		return
	}
}
