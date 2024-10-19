package config

import "github.com/Guise322/TeleBot/server/internal/entities/config"

type Configer interface {
	GetConfig() (config.Config, error)
}
