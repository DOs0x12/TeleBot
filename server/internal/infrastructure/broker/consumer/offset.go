package consumer

import "time"

type offsetWithTimeStamp struct {
	value     int64
	timeStamp time.Time
}

func (kr KafkaConsumer) addOrUpdateOffset(partition int, offset int64) {
	kr.offsetMU.Lock()
	kr.offsets[partition] = offsetWithTimeStamp{value: offset, timeStamp: time.Now()}
	kr.offsetMU.Unlock()
}

func (kr KafkaConsumer) getOffset(partition int) (offsetWithTimeStamp, bool) {
	kr.offsetMU.Lock()
	offset, ok := kr.offsets[partition]
	kr.offsetMU.Unlock()
	return offset, ok
}

func (kr KafkaConsumer) removeOldOffsets(offsets map[int]offsetWithTimeStamp, threashold time.Duration) {
	now := time.Now()
	kr.offsetMU.Lock()
	defer kr.offsetMU.Unlock()
	for part, offset := range offsets {
		if offset.timeStamp.Add(threashold).Before(now) {
			delete(offsets, part)
		}
	}
}
