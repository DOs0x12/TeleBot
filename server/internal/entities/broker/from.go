package broker

import "github.com/google/uuid"

type DataFrom struct {
	IsCommand bool
	Value     string
	MsgUuid   uuid.UUID
}
