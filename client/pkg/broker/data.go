package broker

import "github.com/google/uuid"

type BotData struct {
	ChatID  int64
	Value   string
	msgUuid uuid.UUID
}
