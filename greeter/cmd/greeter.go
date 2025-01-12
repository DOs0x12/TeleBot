package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/DOs0x12/TeleBot/client/v2/broker"

	"github.com/sirupsen/logrus"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	kafkaAddr := flag.String("conn", "kafka:9092", "The kafka connection string.")
	flag.Parse()

	commData := broker.BrokerCommandData{Name: "hello", Description: "Say hello to the bot"}

	serviceName := "greeter"
	br, err := broker.NewKafkaBroker(ctx, *kafkaAddr, serviceName)
	if err != nil {
		logrus.Error(err)

		return
	}

	err = br.RegisterCommand(ctx, commData, serviceName)
	if err != nil {
		logrus.Error("Failed to register a command: ", err)

		return
	}

	inDataChan := br.StartGetData(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case inData := <-inDataChan:
			outData := broker.BrokerData{ChatID: inData.ChatID, Value: "Hello there!"}
			err := br.SendData(ctx, outData)
			if err != nil {
				logrus.Error("Failed to send data: ", err)

				continue
			}
			err = br.Commit(ctx, inData.MessageUuid)
			if err != nil {
				logrus.Error("Failed to commit data: ", err)

				continue
			}
		}
	}
}
