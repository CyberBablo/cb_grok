package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	TelegramBot TelegramBot `yaml:"telegram_bot"`
	Binance     Exchange    `yaml:"binance"`
}

type Exchange struct {
	ProxyUrl  string `yaml:"proxy_url"`
	IsDemo    bool   `yaml:"is_demo"`
	ApiSecret string `yaml:"api_secret"`
	ApuPublic string `yaml:"apu_public"`
}

type TelegramBot struct {
	Token  string `yaml:"token"`
	ChatId int64  `yaml:"chat_id"`
}

func NewConfig() (Config, error) {
	var cfg Config

	file, err := os.Open("config/config.yml")
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
