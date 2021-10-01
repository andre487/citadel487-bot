package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/akamensky/argparse"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	parser := argparse.NewParser("citadel487-bot", "Citadel487 Telegram bot")
	tokenFile := parser.String(
		"t", "token-file",
		&argparse.Options{Required: false, Help: "Tg bot token file"},
	)

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	var token string
	if len(*tokenFile) > 0 {
		val, err := ioutil.ReadFile(*tokenFile)
		if err != nil {
			log.Panic(err)
		}
		token = strings.TrimSpace(string(val))
	} else {
		token = os.Getenv("BOT_TOKEN")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	InitBotActions(bot)
}
