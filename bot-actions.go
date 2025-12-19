package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BotActionsParams struct {
	Bot *tgbotapi.BotAPI
}

func InitBotActions(params BotActionsParams) {
	bot := params.Bot
	if os.Getenv("BOT_DEBUG") == "1" {
		bot.Debug = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		PanicOnErr(err)
	}

	Logger.Info("Waiting for updates")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		onUpdate(params, &update)
	}
}

func onUpdate(params BotActionsParams, update *tgbotapi.Update) {
	message := update.Message
	Logger.Debug("Received message:", message.Chat.ID, message.From.UserName, strings.ReplaceAll(message.Text, "\n", " "))

	var err error
	if message.From.UserName != AllowedUserName || message.Chat.ID != int64(AllowedChat) {
		rickRollMsg := tgbotapi.NewMessage(message.Chat.ID, "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
		_, err = params.Bot.Send(rickRollMsg)
		if err != nil {
			Logger.Warning(fmt.Sprintf("Error sending rickRollMsg: %s", err.Error()))
		}

		alertText := []string{strconv.FormatInt(message.Chat.ID, 10), message.From.UserName, "was rickrolled"}
		alertMsg := tgbotapi.NewMessage(int64(AllowedChat), strings.Join(alertText, " "))
		_, err = params.Bot.Send(alertMsg)
		if err != nil {
			Logger.Warning(fmt.Sprintf("Error sending alertMsg: %s", err.Error()))
		}

		return
	}

	actionDefault(params, update)
}

func actionDefault(params BotActionsParams, update *tgbotapi.Update) {
	message := update.Message

	msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown action")
	msg.ReplyToMessageID = message.MessageID

	_, err := params.Bot.Send(msg)
	if err != nil {
		Logger.Warning(fmt.Sprintf("Error sending msg: %s", err.Error()))
	}
}
