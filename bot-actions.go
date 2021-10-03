package main

import (
	"os"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BotActionsParams struct {
	Bot            *tgbotapi.BotAPI
	S3Endpoint     string
	S3Region       string
	S3Bucket       string
	S3Access       string
	S3Secret       string
	DownloaderPath string
	DownloadDir    string
}

var downloadRegexp = regexp.MustCompile(`(https?://\S+)(?:\s|$)`)

func InitBotActions(params BotActionsParams) error {
	bot := params.Bot
	if os.Getenv("BOT_DEBUG") == "1" {
		bot.Debug = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	Logger.Info("Waiting for updates")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		onUpdate(params, &update)
	}

	return nil
}

func onUpdate(params BotActionsParams, update *tgbotapi.Update) {
	message := update.Message
	Logger.Debug("Received message:", message.From.UserName, strings.ReplaceAll(message.Text, "\n", " "))

	cleanText := strings.TrimSpace(message.Text)
	if strings.HasPrefix(cleanText, "/download") {
		actionDownload(params, update)
	} else {
		actionDefault(params, update)
	}
}

func actionDownload(params BotActionsParams, update *tgbotapi.Update) {
	message := update.Message

	urls := downloadRegexp.FindAllString(message.Text, -1)
	if urls == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "No URLs found in message")
		msg.ReplyToMessageID = message.MessageID

		params.Bot.Send(msg)
		return
	}

	dwlParams := DownloadByUrlParams{
		Message: message,
		Urls:    urls,
	}
	dwlParams.Bot = params.Bot
	dwlParams.S3Endpoint = params.S3Endpoint
	dwlParams.S3Region = params.S3Region
	dwlParams.S3Bucket = params.S3Bucket
	dwlParams.S3Access = params.S3Access
	dwlParams.S3Secret = params.S3Secret
	dwlParams.DownloaderPath = params.DownloaderPath
	dwlParams.DownloadDir = params.DownloadDir

	go DownloadByUrl(dwlParams)
}

func actionDefault(params BotActionsParams, update *tgbotapi.Update) {
	message := update.Message

	msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown action")
	msg.ReplyToMessageID = message.MessageID

	params.Bot.Send(msg)
}
