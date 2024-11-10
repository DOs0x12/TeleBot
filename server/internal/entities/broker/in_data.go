package broker

import "github.com/google/uuid"

type InData struct {
	IsCommand bool
	Value     string
	MsgUuid   uuid.UUID
}
