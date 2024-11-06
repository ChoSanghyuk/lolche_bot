package config

import (
	_ "embed"
	"strconv"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var configByte []byte

type Config struct {
	TeleBot struct {
		Token  string `yaml:"token"`
		ChatId string `yaml:"chatId"`
	} `yaml:"telegram"`
}

func NewConfig() (*Config, error) {
	var ConfigInfo Config = Config{}

	err := yaml.Unmarshal(configByte, &ConfigInfo)
	if err != nil {
		return nil, err
	}
	return &ConfigInfo, nil
}

func (c Config) Telebot() (token string, chatId int64) {
	chatId, _ = strconv.ParseInt(c.TeleBot.ChatId, 10, 64)
	return c.TeleBot.Token, chatId
}
