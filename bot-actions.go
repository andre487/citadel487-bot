package main

import (
	"os"
	"regexp"
	"strconv"
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

// TODO: move to config
var allowedUserName = "andre487"
var allowedChat = 94764326

func InitBotActions(params BotActionsParams) error {
	PrepareNetRc()

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

	downloadChannel := make(chan DownloadByUrlParams, 2048)
	go DownloadByUrlWithQueue(&downloadChannel)

	Logger.Info("Waiting for updates")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		onUpdate(params, &downloadChannel, &update)
	}

	return nil
}

func onUpdate(params BotActionsParams, downloadChannel *chan DownloadByUrlParams, update *tgbotapi.Update) {
	message := update.Message
	Logger.Debug("Received message:", message.Chat.ID, message.From.UserName, strings.ReplaceAll(message.Text, "\n", " "))

	if message.From.UserName != allowedUserName || message.Chat.ID != int64(allowedChat) {
		rickRollMsg := tgbotapi.NewMessage(message.Chat.ID, "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
		params.Bot.Send(rickRollMsg)

		alertText := []string{strconv.FormatInt(message.Chat.ID, 10), message.From.UserName, "was rickrolled"}
		alertMsg := tgbotapi.NewMessage(int64(allowedChat), strings.Join(alertText, " "))
		params.Bot.Send(alertMsg)

		return
	}

	cleanText := strings.TrimSpace(message.Text)
	if strings.HasPrefix(cleanText, "/download") {
		actionDownloadUrl(params, downloadChannel, update)
	} else if message.Document != nil {
		DownloadDocument(params, update)
	} else if message.Photo != nil {
		DownloadPhoto(params, update)
	} else if message.Video != nil {
		DownloadVideo(params, update)
	} else {
		actionDefault(params, update)
	}
}

func actionDownloadUrl(params BotActionsParams, downloadChannel *chan DownloadByUrlParams, update *tgbotapi.Update) {
	message := update.Message

	urls := downloadRegexp.FindAllString(message.Text, -1)
	if urls == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "No URLs found in message")
		msg.ReplyToMessageID = message.MessageID

		params.Bot.Send(msg)
		return
	}

	downloadFiles(params, downloadChannel, message, urls)
}

func actionDefault(params BotActionsParams, update *tgbotapi.Update) {
	message := update.Message

	msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown action")
	msg.ReplyToMessageID = message.MessageID

	params.Bot.Send(msg)
}

func downloadFiles(
	params BotActionsParams, downloadChannel *chan DownloadByUrlParams,
	message *tgbotapi.Message, urls []string,
) {
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

	*downloadChannel <- dwlParams
}
