package telegram

import (
	"cb_grok/config"
	"errors"
	"fmt"
	tele "github.com/tucnak/telebot"
	"log"
)

type TelegramService struct {
	bot    *tele.Bot
	chatID int64
}

func NewTelegramService(cfg config.Config) (*TelegramService, error) {
	bot, err := tele.NewBot(tele.Settings{
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
		_, err := s.bot.Send(&tele.Chat{ID: s.chatID}, text)
		if err != nil {
			log.Printf("telegram send error: %s\n", err.Error())
		}
	}()
}

func (s *TelegramService) SendFile(path string, text string) {
	go func() {
		fmt.Println(path)
		doc := &tele.Document{
			File:     tele.FromDisk(path),
			Caption:  text,
			FileName: "events.html"}
		_, err := s.bot.Send(&tele.Chat{ID: s.chatID}, doc)
		if err != nil {
			log.Printf("telegram send error: %s\n", err.Error())
		}
	}()
}
