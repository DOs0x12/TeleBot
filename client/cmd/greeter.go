package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/Guise322/TeleBot/client/pkg/broker"

	"github.com/sirupsen/logrus"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	kafkaAddr := flag.String("conn", "kafka:9092", "The kafka connection string.")
	flag.Parse()

	commData := broker.CommandData{Name: "/hello", Description: "Say hello to the bot"}
	r := broker.NewReceiver(*kafkaAddr, commData.Name)
	s := broker.NewSender(*kafkaAddr)

	err := s.RegisterCommand(ctx, commData)
	if err != nil {
		logrus.Error("Failed to register a command: ", err)

		return
	}

	inDataChan := r.StartGetData(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case inData := <-inDataChan:
			outData := broker.BotData{ChatID: inData.ChatID, Value: "Hello there!"}
			s.SendData(ctx, outData)
		}
	}
}
