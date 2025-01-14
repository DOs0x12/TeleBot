package topic

import (
	"fmt"
	"net"
	"strconv"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func CreateDataTopic(topicName, address string) error {
	conn, err := kafka.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to the broker with the address %v: %w", address, err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logrus.Error("Failed to close the broker connection: ", err)
		}
	}()

	contr, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get the controller URL: %w", err)
	}

	contrAddr := net.JoinHostPort(contr.Host, strconv.Itoa(contr.Port))
	contrConn, err := kafka.Dial("tcp", contrAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to the controller with address %v: %w", contrAddr, err)
	}

	defer func() {
		if err := contrConn.Close(); err != nil {
			logrus.Error("Failed to close the controller connection: ", err)
		}
	}()

	topicConfigs := []kafka.TopicConfig{{Topic: topicName, NumPartitions: 1, ReplicationFactor: 1}}
	err = contrConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create a topic with the name %v: %w", topicName, err)
	}

	return nil
}
