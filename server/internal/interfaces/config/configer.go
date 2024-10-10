package config

import "TeleBot/internal/entities/config"

type Configer interface {
	GetConfig() (config.Config, error)
}
