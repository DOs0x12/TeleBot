package consumer

import "time"

type offsetWithTimeStamp struct {
	value     int64
	timeStamp time.Time
}

func (kr *KafkaConsumer) addOrUpdateOffset(partition int, offset int64) {
	kr.offsetMU.Lock()
	kr.offsets[partition] = offsetWithTimeStamp{value: offset, timeStamp: time.Now()}
	kr.offsetMU.Unlock()
}

func (kr *KafkaConsumer) getOffset(partition int) (offsetWithTimeStamp, bool) {
	kr.offsetMU.Lock()
	offset, ok := kr.offsets[partition]
	kr.offsetMU.Unlock()
	return offset, ok
}
