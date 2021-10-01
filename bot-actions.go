package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func InitBotActions(bot *tgbotapi.BotAPI) {
	if os.Getenv("BOT_DEBUG") == "1" {
		bot.Debug = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Waiting for updates")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		onUpdate(bot, &update)
	}
}

func onUpdate(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	message := update.Message
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	cleanText := strings.TrimSpace(message.Text)
	if strings.HasPrefix(cleanText, "/download") {
		actionDownload(bot, update)
	} else {
		actionDefault(bot, update)
	}
}

func actionDownload(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	message := update.Message

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Will download %s", message.Text))
	msg.ReplyToMessageID = message.MessageID

	bot.Send(msg)
}

func actionDefault(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	message := update.Message

	msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown action")
	msg.ReplyToMessageID = message.MessageID

	bot.Send(msg)
}
