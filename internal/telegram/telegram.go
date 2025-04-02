package telegram

import (
	"cb_grok/config"
	"errors"
	"github.com/tucnak/telebot"
	"log"
)

type TelegramService struct {
	bot    *telebot.Bot
	chatID int64
}

func NewTelegramService(cfg config.Config) (*TelegramService, error) {
	bot, err := telebot.NewBot(telebot.Settings{
		Token: cfg.TelegramBot.Token,
	})
	if err != nil {
		return nil, err
	}

	if cfg.TelegramBot.ChatId == 0 {
		return nil, errors.New("chat ID not provided")
	}
	chatID := cfg.TelegramBot.ChatId

	return &TelegramService{
		bot:    bot,
		chatID: chatID,
	}, nil
}

func (s *TelegramService) SendMessage(text string) {
	go func() {
		_, err := s.bot.Send(&telebot.Chat{ID: s.chatID}, text)
		if err != nil {
			log.Printf("telegram send error: %s\n", err.Error())
		}
	}()
}
