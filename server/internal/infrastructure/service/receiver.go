package service

import (
	"TeleBot/internal/entities/service"
	"context"
	"fmt"
	"net"
	"strconv"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type KafkaReceiver struct {
	address string
	port    string
}

func NewKafkaReceiver(address, port string) KafkaReceiver {
	return KafkaReceiver{address: address, port: port}
}

func (kr KafkaReceiver) StartReceivingData(ctx context.Context) <-chan service.InData {
	kr.createDataTopic()
	dataChan := make(chan service.InData)

	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:     "regdfgd1",
		Brokers:     []string{fmt.Sprintf("%v:%v", kr.address, kr.port)},
		Topic:       "botdata",
		Partition:   0,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	go consumeMessages(ctx, dataChan, r)

	return dataChan
}

func (kr KafkaReceiver) createDataTopic() {
	conn, err := kafka.Dial("tcp", fmt.Sprintf("%v:%v", kr.address, kr.port))
	if err != nil {
		panic(err.Error())
	}

	defer conn.Close()

	contr, err := conn.Controller()
	if err != nil {
		panic(err.Error())
	}

	contrConn, err := kafka.Dial("tcp", net.JoinHostPort(contr.Host, strconv.Itoa(contr.Port)))
	if err != nil {
		panic(err.Error())
	}

	defer contrConn.Close()

	topicConfigs := []kafka.TopicConfig{{Topic: "botdata", NumPartitions: 1, ReplicationFactor: 1}}
	err = contrConn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}
}

func consumeMessages(ctx context.Context, dataChan chan service.InData, r *kafka.Reader) {
	for {
		if ctx.Err() != nil {
			break
		}

		msg, err := r.FetchMessage(ctx)
		if err != nil {
			continue
		}

		commandKey := "command"
		isCommand := string(msg.Key) == commandKey

		dataChan <- service.InData{IsCommand: isCommand, Value: string(msg.Value)}
		r.CommitMessages(ctx, msg)
	}

	if err := r.Close(); err != nil {
		logrus.Fatal("failed to close reader:", err)
	}
}
