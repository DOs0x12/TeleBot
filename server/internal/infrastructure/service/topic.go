package service

import (
	"net"
	"strconv"

	kafka "github.com/segmentio/kafka-go"
)

func createDataTopic(topicName, address string) {
	conn, err := kafka.Dial("tcp", address)
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

	topicConfigs := []kafka.TopicConfig{{Topic: topicName, NumPartitions: 1, ReplicationFactor: 1}}
	err = contrConn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}
}
