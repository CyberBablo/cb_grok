package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	TelegramBot TelegramBot `yaml:"telegram_bot"`
	Binance     Exchange    `yaml:"binance"`
	PostgreSQL  PostgreSQL  `yaml:"postgresql"`
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

type PostgreSQL struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

func (p *PostgreSQL) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User,
		p.Password,
		p.Host,
		p.Port,
		p.Database,
		p.SSLMode,
	)
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
