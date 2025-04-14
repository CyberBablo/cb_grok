package telegram

import (
	"bytes"
	"cb_grok/config"
	"errors"
	"fmt"
	tele "github.com/tucnak/telebot"
	"log"
	"os"
	"strings"
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

//func (s *TelegramService) SendFile(b *bytes.Buffer, text string) {
//	go func() {
//
//		doc := &tele.Document{
//			File:     tele.FromDisk(path),
//			Caption:  text,
//			FileName: "events.html"}
//
//		_, err := s.bot.Send(&tele.Chat{ID: s.chatID}, doc)
//		if err != nil {
//			log.Printf("telegram send error: %s\n", err.Error())
//		}
//	}()
//}

func (s *TelegramService) SendFile(b *bytes.Buffer, fileExtension string, text string) error {
	// Validate input
	if b == nil || b.Len() == 0 {
		return fmt.Errorf("empty byte buffer provided")
	}
	if fileExtension == "" {
		fileExtension = "txt" // Default to .txt if no extension provided
	}

	// Normalize file extension (remove leading dot if present)
	fileExtension = strings.TrimPrefix(fileExtension, ".")
	// Create a temporary file with appropriate extension
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("telegram-file-*.%s", fileExtension))
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write bytes to temporary file
	if _, err := tmpFile.Write(b.Bytes()); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Send file using telebot
	go func() {
		defer os.Remove(tmpFile.Name()) // Ensure file is deleted after function ends

		doc := &tele.Document{
			File:    tele.FromDisk(tmpFile.Name()),
			Caption: text,
		}

		_, err := s.bot.Send(&tele.Chat{ID: s.chatID}, doc)
		if err != nil {
			log.Printf("telegram send error: %s\n", err.Error())
		}
	}()

	return nil
}
