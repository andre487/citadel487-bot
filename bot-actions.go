package main

import (
	"log"
	"os"

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

		message := update.Message
		log.Printf("[%s] %s", message.From.UserName, message.Text)

		msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)
		msg.ReplyToMessageID = message.MessageID

		bot.Send(msg)
	}
}
