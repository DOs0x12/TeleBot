package service

import (
	"context"
	"strings"
	"time"

	"github.com/Guise322/TeleBot/server/internal/common"
	"github.com/Guise322/TeleBot/server/internal/entities/service"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaTransmitter struct {
	w *kafka.Writer
}

var lastCommand string

func NewKafkaTransmitter(address string) KafkaTransmitter {
	w := &kafka.Writer{
		Addr:     kafka.TCP(address),
		Balancer: &kafka.LeastBytes{},
	}

	return KafkaTransmitter{w: w}
}

func (kt KafkaTransmitter) TransmitData(ctx context.Context, data service.OutData) {
	retryNum := 10
	waitTime := 500 * time.Millisecond

	for i := 0; i < retryNum; i++ {
		if ctx.Err() != nil {
			break
		}

		if lastCommand == "" && data.CommName == "" {
			logrus.Warn("Get an empty message")

			continue
		}

		if data.CommName == "" {
			data.CommName = lastCommand
		}

		topicName := strings.Trim(data.CommName, "/")

		err := sendMessage(ctx, kt.w, topicName, data.Value)
		if err != nil {
			if err == kafka.UnknownTopicOrPartition {
				logrus.WithField("topiName", topicName).Warn("An unknown topic, create the one")
				createDataTopic(topicName, kt.w.Addr.String())
				err = sendMessage(ctx, kt.w, topicName, data.Value)
			}
		}

		if err == nil {
			if data.CommName != "" {
				lastCommand = data.CommName
			}

			return
		}

		logrus.Error("Failed to write messages:", err)

		common.WaitWithContext(ctx, waitTime)
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

func (kt KafkaTransmitter) Close() {
	if err := kt.w.Close(); err != nil {
		logrus.Error("Failed to close writer:", err)
	}
}
