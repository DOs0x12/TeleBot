package service

import (
	"TeleBot/internal/entities/service"
	"context"
)

type DataReceiver interface {
	StartReceivingData(ctx context.Context) <-chan service.InData
}
