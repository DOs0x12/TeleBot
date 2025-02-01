package consumer

import (
	"time"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

type uncommittedMessage struct {
	msg       kafka.Message
	timeStamp time.Time
}

func (kr *KafkaConsumer) addMsgToUncommitted(msg kafka.Message) uuid.UUID {
	msgUuid := uuid.New()
	kr.uncomMsgMU.Lock()
	kr.uncommittedMessages[msgUuid] = uncommittedMessage{msg: msg, timeStamp: time.Now()}
	kr.uncomMsgMU.Unlock()

	return msgUuid
}

func (kr *KafkaConsumer) delMsgFromUncommitted(msgUuid uuid.UUID) {
	kr.uncomMsgMU.Lock()
	delete(kr.uncommittedMessages, msgUuid)
	kr.uncomMsgMU.Unlock()
}

func (kr *KafkaConsumer) getMsgFromUncommited(msgUuid uuid.UUID) (uncommittedMessage, bool) {
	kr.uncomMsgMU.Lock()
	msg, ok := kr.uncommittedMessages[msgUuid]
	kr.uncomMsgMU.Unlock()
	return msg, ok
}
