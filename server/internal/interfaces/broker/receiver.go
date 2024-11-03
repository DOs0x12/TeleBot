package service

import (
	"context"

	"github.com/Guise322/TeleBot/server/internal/entities/service"
)

type DataReceiver interface {
	StartReceivingData(ctx context.Context) (<-chan service.InData, error)
}
