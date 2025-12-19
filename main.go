package main

import (
	"os"

	"github.com/akamensky/argparse"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	SetupLogger()
	parser := argparse.NewParser("citadel487-bot", "Citadel487 Telegram bot")

	sqsEndpoint := parser.String("", "sqs-endpoint", &argparse.Options{Default: "https://message-queue.api.cloud.yandex.net"})
	sqsRegion := parser.String("", "sqs-region", &argparse.Options{Default: "ru-central1"})

	err := parser.Parse(os.Args)
	PanicOnErr(err)

	secretProvider := NewSecretProvider()

	token := secretProvider.BotToken()

	bot, err := tgbotapi.NewBotAPI(token)
	PanicOnErr(err)

	go ReceiveSms(bot, *sqsEndpoint, *sqsRegion, secretProvider.SqsParams())

	InitBotActions(BotActionsParams{
		Bot: bot,
	})
}
