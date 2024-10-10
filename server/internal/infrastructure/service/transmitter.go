package service

import (
	"TeleBot/internal/entities/service"
	"context"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaTransmitter struct {
	address string
	port    string
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
			err := w.WriteMessages(ctx,
				kafka.Message{
					Topic: data.CommName,
					Value: []byte(data.Value),
				},
			)
			if err != nil {
				logrus.Fatal("failed to write messages:", err)
			}

			if err := w.Close(); err != nil {
				logrus.Fatal("failed to close writer:", err)
			}
		}
	}
}
