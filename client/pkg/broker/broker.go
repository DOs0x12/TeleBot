package broker

import "github.com/segmentio/kafka-go"

type BotData struct {
	ChatID int64
	Value  string
}

type Broker struct {
	w *kafka.Writer
}

func NewBroker(address string) Broker {
	w := &kafka.Writer{
		Addr:     kafka.TCP(address),
		Balancer: &kafka.LeastBytes{},
	}

	return Broker{w: w}
}
