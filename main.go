package main

import (
	"os"

	"github.com/akamensky/argparse"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	SetupLogger()
	parser := argparse.NewParser("citadel487-bot", "Citadel487 Telegram bot")

	// secretProvider := NewSecretProvider()
	// Logger.Info(secretProvider.BotToken())
	// Logger.Info(secretProvider.S3Params())
	// Logger.Info(secretProvider.SqsParams())

	tokenFile := parser.String("t", "token-file", &argparse.Options{Required: false, Help: "Tg bot token file", Default: "~/.tokens/dev-tg-bot"})
	s3Endpoint := parser.String("", "s3-endpoint", &argparse.Options{Default: "https://storage.yandexcloud.net"})
	s3Region := parser.String("", "s3-region", &argparse.Options{Default: "ru-central1"})
	s3Bucket := parser.String("", "s3-bucket", &argparse.Options{Default: "downloader487-files"})
	s3AccessFile := parser.String("", "s3-access-file", &argparse.Options{Default: "~/.tokens/s3-access"})
	s3SecretFile := parser.String("", "s3-secret-file", &argparse.Options{Default: "~/.tokens/s3-secret"})
	downloaderPath := parser.String("", "downloader-path", &argparse.Options{Default: "./downloader487/downloader/dist/downloader487"})
	downloadDir := parser.String("", "download-dir", &argparse.Options{Default: "/tmp/citadel487-bot/downloads"})

	err := parser.Parse(os.Args)
	PanicOnErr(err)

	token, err := GetSecretValue("BOT_TOKEN", tokenFile)
	PanicOnErr(err)
	s3Access, err := GetSecretValue("S3_ACCESS_KEY", s3AccessFile)
	PanicOnErr(err)
	s3Secret, err := GetSecretValue("S3_SECRET_KEY", s3SecretFile)
	PanicOnErr(err)

	bot, err := tgbotapi.NewBotAPI(token)
	PanicOnErr(err)

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
