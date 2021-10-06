package main

import (
	"os"

	"github.com/akamensky/argparse"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	SetupLogger()
	parser := argparse.NewParser("citadel487-bot", "Citadel487 Telegram bot")

	s3Endpoint := parser.String("", "s3-endpoint", &argparse.Options{Default: "https://storage.yandexcloud.net"})
	s3Region := parser.String("", "s3-region", &argparse.Options{Default: "ru-central1"})
	s3Bucket := parser.String("", "s3-bucket", &argparse.Options{Default: "downloader487-files"})
	sqsEndpoint := parser.String("", "sqs-endpoint", &argparse.Options{Default: "https://message-queue.api.cloud.yandex.net"})
	sqsRegion := parser.String("", "sqs-region", &argparse.Options{Default: "ru-central1"})
	downloaderPath := parser.String("", "downloader-path", &argparse.Options{Default: "./downloader487/downloader/dist/downloader487"})
	downloadDir := parser.String("", "download-dir", &argparse.Options{Default: "/tmp/citadel487-bot/downloads"})

	err := parser.Parse(os.Args)
	PanicOnErr(err)

	secretProvider := NewSecretProvider()

	token := secretProvider.BotToken()
	s3Data := secretProvider.S3Params()
	s3Access := s3Data.AccessKey
	s3Secret := s3Data.SecretKey

	bot, err := tgbotapi.NewBotAPI(token)
	PanicOnErr(err)

	go ReceiveSms(bot, *sqsEndpoint, *sqsRegion, secretProvider.SqsParams())

	InitBotActions(BotActionsParams{
		Bot:            bot,
		S3Endpoint:     *s3Endpoint,
		S3Region:       *s3Region,
		S3Bucket:       *s3Bucket,
		S3Access:       s3Access,
		S3Secret:       s3Secret,
		DownloaderPath: *downloaderPath,
		DownloadDir:    *downloadDir,
	})
}
