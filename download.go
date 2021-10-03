package main

import (
	"bytes"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type DownloadByUrlParams struct {
	Message *tgbotapi.Message
	Urls    []string
	BotActionsParams
}

func DownloadByUrlWithQueue(channel *chan DownloadByUrlParams) {
	for params := range *channel {
		DownloadByUrl(params)
	}
}

func DownloadByUrl(params DownloadByUrlParams) {
	args := []string{
		"--download-dir", params.DownloadDir,
		"--s3-endpoint", params.S3Endpoint,
		"--s3-region", params.S3Region,
		"--s3-bucket", params.S3Bucket,
	}

	for _, url := range params.Urls {
		args = append(args, "--url", strings.TrimSpace(url))
	}

	cmd := exec.Command(params.DownloaderPath, args...)
	cmd.Env = append(
		cmd.Env,
		"LC_ALL=C.UTF-8",
		"LANG=C.UTF-8",
		"S3_ACCESS_KEY="+params.S3Access,
		"S3_SECRET_KEY="+params.S3Secret,
	)

	var cmdStderr bytes.Buffer
	cmd.Stderr = &cmdStderr

	Logger.Debug("Run downloader:", cmd)
	startMsg := tgbotapi.NewMessage(params.Message.Chat.ID, "Download started")
	startMsg.ReplyToMessageID = params.Message.MessageID
	params.Bot.Send(startMsg)

	err := cmd.Run()
	Logger.Debug("Downloader stderr:", cmdStderr.String())

	res := []string{"Download result is "}
	if err == nil {
		res = append(res, "success")
	} else {
		res = append(res, "error: "+err.Error())
	}

	msgText := strings.Join(res, "")
	Logger.Debug(msgText)

	logFileBytes := tgbotapi.FileBytes{
		Name:  "download.txt",
		Bytes: cmdStderr.Bytes(),
	}
	logFile := tgbotapi.NewDocumentUpload(params.Message.Chat.ID, logFileBytes)
	logFile.ReplyToMessageID = params.Message.MessageID
	logFile.Caption = msgText
	_, sendLogErr := params.Bot.Send(logFile)
	if sendLogErr != nil {
		Logger.Error("Send log error:", sendLogErr.Error())
	}
}
