package config

import (
	_ "embed"
	t "lolcheBot"
	"lolcheBot/db"
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

	Db struct {
		User     string `yaml:"user"`
		Password string `yaml:"pw"`
		IP       string `yaml:"ip"`
		Port     string `yaml:"port"`
		Scheme   string `yaml:"scheme"`
	} `yaml:"db"`
}

func NewConfig() (*Config, error) {
	var ConfigInfo Config = Config{}

	err := yaml.Unmarshal(configByte, &ConfigInfo)
	if err != nil {
		return nil, err
	}
	return &ConfigInfo, nil
}

func (c Config) Telebot() *t.TeleBotConfig {
	chatId, _ := strconv.ParseInt(c.TeleBot.ChatId, 10, 64)
	return t.NewTeleBotConfig(c.TeleBot.Token, chatId)
}

func (c Config) StorageConfig() *db.StorageConfig {
	return db.NewStorageConfig(
		c.Db.User,
		c.Db.Password,
		c.Db.IP,
		c.Db.Port,
		c.Db.Scheme,
	)

}
