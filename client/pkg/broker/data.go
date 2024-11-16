package broker

import "github.com/google/uuid"

type BotData struct {
	ChatID      int64
	Value       string
	MessageUuid uuid.UUID
}
