package config

import (
	"fmt"
	"os"

	"github.com/DOs0x12/TeleBot/server/internal/entities/config"

	"gopkg.in/yaml.v3"
)

type ConfigDto struct {
	BotKey       string `yaml:"botKey"`
	KafkaAddress string `yaml:"kafkaAddress"`
}

type Configer struct {
	path string
}

func (c Configer) LoadConfig() (config.Config, error) {
	confFile, err := os.ReadFile(c.path)
	if err != nil {
		return config.Config{}, fmt.Errorf("an error of reading the config file: %w", err)
	}

	return unmarshalConf(confFile)
}

func NewConfiger(path string) Configer {
	return Configer{path: path}
}

func unmarshalConf(data []byte) (config.Config, error) {
	dto := ConfigDto{}
	err := yaml.Unmarshal(data, &dto)
	if err != nil {
		return config.Config{}, fmt.Errorf("an error of unmarshalling config data: %w", err)
	}

	return cast(dto), nil
}

func cast(confDto ConfigDto) config.Config {
	return config.Config{BotKey: confDto.BotKey, KafkaAddress: confDto.KafkaAddress}
}
