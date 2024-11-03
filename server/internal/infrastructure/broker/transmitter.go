package service

import (
	"context"
	"strings"

	"github.com/Guise322/TeleBot/server/internal/entities/service"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaTransmitter struct {
	address string
}

var lastCommand string

func NewKafkaTransmitter(address string) KafkaTransmitter {
	return KafkaTransmitter{address: address}
}

func (kt KafkaTransmitter) StartTransmittingData(ctx context.Context) chan<- service.OutData {
	dataChan := make(chan service.OutData)

	w := &kafka.Writer{
		Addr:     kafka.TCP(kt.address),
		Balancer: &kafka.LeastBytes{},
	}

	go transmitData(ctx, dataChan, w)

	return dataChan
}

func transmitData(ctx context.Context, dataChan chan service.OutData, w *kafka.Writer) {
	for {
		select {
		case <-ctx.Done():
			if err := w.Close(); err != nil {
				logrus.Error("Failed to close writer:", err)
			}
		case data := <-dataChan:
			if lastCommand == "" && data.CommName == "" {
				logrus.Warnf("Get an empty message")

				continue
			}

			topicName := strings.Trim(data.CommName, "/")

			err := sendMessage(ctx, w, topicName, data.Value)
			if err != nil {
				if err == kafka.UnknownTopicOrPartition {
					createDataTopic(topicName, w.Addr.String())
					err = sendMessage(ctx, w, topicName, data.Value)
				}

				if err != nil {
					logrus.Error("Failed to write messages:", err)
				}

				continue
			}

			if data.CommName != "" {
				lastCommand = data.CommName
			}
		}
	}
}

func sendMessage(ctx context.Context, w *kafka.Writer, commName, data string) error {
	return w.WriteMessages(ctx,
		kafka.Message{
			Topic: commName,
			Value: []byte(data),
		},
	)
}
