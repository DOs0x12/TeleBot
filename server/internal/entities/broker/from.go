package broker

import "github.com/google/uuid"

type DataFrom struct {
	ChatID  int64
	Value   []byte
	MsgUuid uuid.UUID
	IsFile  bool
}

type CommandFrom struct {
	Name,
	Description,
	Token string
	MsgUuid uuid.UUID
}
