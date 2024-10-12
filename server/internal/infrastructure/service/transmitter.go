package service

import (
	"TeleBot/internal/entities/service"
	"context"
	"strings"

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
			return
		case data := <-dataChan:
			if lastCommand == "" && data.CommName == "" {
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
					logrus.Error("failed to write messages:", err)
				}
			}

			if err := w.Close(); err != nil {
				logrus.Error("failed to close writer:", err)

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
