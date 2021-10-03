package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type DownloadByUrlParams struct {
	Message *tgbotapi.Message
	Urls    []string
	BotActionsParams
}

func PrepareNetRc() {
	netRcContent := os.Getenv("NETRC")
	if len(netRcContent) == 0 {
		return
	}
	home := os.Getenv("HOME")
	ioutil.WriteFile(path.Join(home, ".netrc"), []byte(netRcContent), 0600)
	Logger.Info("Netrc file of", len(netRcContent), "bytes has been written")
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

func DownloadDocument(params BotActionsParams, update *tgbotapi.Update) {
	message := update.Message
	Logger.Debug("Got document", message.Document.FileID)

	url, err := params.Bot.GetFileDirectURL(message.Document.FileID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Can't get URL of document: "+err.Error())
		msg.ReplyToMessageID = message.MessageID
		params.Bot.Send(msg)
		Logger.Error("URL error:", err.Error())
		return
	}
	Logger.Debug("Got document URL")

	resp, err := http.Get(url)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Can't download document: "+err.Error())
		msg.ReplyToMessageID = message.MessageID
		params.Bot.Send(msg)
		Logger.Error("Download error:", err.Error())
		return
	}

	fileName := strings.ReplaceAll(fmt.Sprintf("telegram-%d-%s", rand.Int(), message.Document.FileName), " ", "_")
	Logger.Debug("File name:", fileName)

	s3Session := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(params.S3Endpoint),
		Region:      aws.String(params.S3Region),
		Credentials: credentials.NewStaticCredentials(params.S3Access, params.S3Secret, ""),
	}))
	uploader := s3manager.NewUploader(s3Session)

	uploadResult, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(params.S3Bucket),
		Key:    aws.String(fileName),
		Body:   resp.Body,
	})
	resp.Body.Close()

	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Can't upload document: "+err.Error())
		msg.ReplyToMessageID = message.MessageID
		params.Bot.Send(msg)
		Logger.Error("Upload error:", err.Error())
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Download success")
	msg.ReplyToMessageID = message.MessageID
	params.Bot.Send(msg)
	Logger.Info("File uploaded to", uploadResult.Location)
}
