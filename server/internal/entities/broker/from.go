package broker

import "github.com/google/uuid"

type DataFrom struct {
	ChatID  int64
	Value   string
	MsgUuid uuid.UUID
}

type CommandFrom struct {
	Name,
	Description,
	Token string
	MsgUuid uuid.UUID
}
