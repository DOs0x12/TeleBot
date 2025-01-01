package config

import (
	"fmt"
	"os"

	"github.com/DOs0x12/TeleBot/server/internal/entities/config"

	"gopkg.in/yaml.v3"
)

type ConfigDto struct {
	BotKey         string `yaml:"botKey"`
	KafkaAddress   string `yaml:"kafkaAddress"`
	StorageAddress string `yaml:"storageAddress"`
	StorageDB      string `yaml:"storageDB"`
	StorageUser    string `yaml:"storageUser"`
	StoragePass    string `yaml:"storagePass"`
}

type Configer struct {
	path string
}

func (c Configer) LoadConfig() (config.Config, error) {
	confFile, err := os.ReadFile(c.path)
	if err != nil {
		return config.Config{}, fmt.Errorf("failed to read the config file: %w", err)
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
		return config.Config{}, fmt.Errorf("failed to unmarshal the config data: %w", err)
	}

	return cast(dto), nil
}

func cast(confDto ConfigDto) config.Config {
	return config.Config{BotKey: confDto.BotKey, KafkaAddress: confDto.KafkaAddress}
}
