package topic

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	kafka "github.com/segmentio/kafka-go"
)

func CreateDataTopic(topicName, address string) (err error) {
	conn, err := kafka.Dial("tcp", address)
	if err != nil {
		err = fmt.Errorf("failed to connect to the broker with the address %v: %w", address, err)

		return
	}

	defer func() {
		if clErr := conn.Close(); clErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close the broker connection: %w", clErr))
		}
	}()

	contr, err := conn.Controller()
	if err != nil {
		err = fmt.Errorf("failed to get the controller URL: %w", err)

		return
	}

	contrAddr := net.JoinHostPort(contr.Host, strconv.Itoa(contr.Port))
	contrConn, err := kafka.Dial("tcp", contrAddr)
	if err != nil {
		err = fmt.Errorf("failed to connect to the controller with address %v: %w", contrAddr, err)

		return
	}

	defer func() {
		if clErr := contrConn.Close(); clErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close the controller connection: %v", clErr))
		}
	}()

	topicConfigs := []kafka.TopicConfig{{Topic: topicName, NumPartitions: 1, ReplicationFactor: 1}}
	err = contrConn.CreateTopics(topicConfigs...)
	if err != nil {
		err = fmt.Errorf("failed to create a topic with the name %v: %w", topicName, err)

		return
	}

	return
}
