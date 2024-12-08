package broker

import "github.com/google/uuid"

// Stores the data that is needed by the bot app.
type BotData struct {
	ChatID      int64
	Value       string
	MessageUuid uuid.UUID
}
