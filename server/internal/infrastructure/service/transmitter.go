package service

import (
	"TeleBot/internal/entities/service"
	"context"
	"encoding/json"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaTransmitter struct {
	address string
	port    string
}

type ServiceOutDataDto struct {
	ChatID int64
	Value  string
}

func NewKafkaTransmitter(address, port string) KafkaTransmitter {
	return KafkaTransmitter{address: address, port: port}
}

func (kt KafkaTransmitter) StartTransmittingData(ctx context.Context) chan<- service.OutData {
	dataChan := make(chan service.OutData)

	w := &kafka.Writer{
		Addr:     kafka.TCP(fmt.Sprintf("%v:%v", kt.address, kt.port)),
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
			msg, err := json.Marshal(ServiceOutDataDto{ChatID: data.ChatID, Value: data.Value})
			if err != nil {
				logrus.Error("failed to marshal an object to message:", err)
			}

			err = w.WriteMessages(ctx,
				kafka.Message{
					Topic: data.CommName,
					Value: msg,
				},
			)
			if err != nil {
				logrus.Error("failed to write messages:", err)
			}

			if err := w.Close(); err != nil {
				logrus.Error("failed to close writer:", err)
			}
		}
	}
}
