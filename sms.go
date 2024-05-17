package main

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type MessageData struct {
	Type string `json:"type"`
	Data []struct {
		MessageType          string `json:"message_type"`
		PrintableMessageType string `json:"printable_message_type"`
		DeviceId             string `json:"device_id"`
		Tel                  string `json:"tel"`
		DateTime             string `json:"date_time"`
		PrintableDateTime    string `json:"printable_date_time"`
		SmsDateTime          string `json:"sms_date_time"`
		Marked               bool   `json:"marked"`
		Text                 string `json:"text"`
	} `json:"data"`
}

var messageTemplate = template.Must(template.New("msg").Parse(`{{define "T"}}
SMS487: <b>{{.Tel}}</b>
<b>{{.PrintableMessageType}} {{.DeviceId}} {{.PrintableDateTime}}</b>
{{.Text}}
{{end}}`))

func ReceiveSms(bot *tgbotapi.BotAPI, sqsEndpoint string, sqsRegion string, sqsParams SqsParamsData) {
	sess := session.Must(session.NewSession())
	svc := sqs.New(
		sess,
		aws.NewConfig().WithEndpoint(
			sqsEndpoint,
		).WithRegion(
			sqsRegion,
		).WithCredentials(
			credentials.NewStaticCredentials(sqsParams.AccessKey, sqsParams.SecretKey, ""),
		),
	)

	Logger.Info("Start to listen SMS queue", sqsParams.QueueUrl)
	for {
		result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			MaxNumberOfMessages: aws.Int64(10),
			QueueUrl:            aws.String(sqsParams.QueueUrl),
			VisibilityTimeout:   aws.Int64(60),
			WaitTimeSeconds:     aws.Int64(20),
		})

		if err != nil {
			msgText := "Error when receiving SQS messages: " + err.Error()
			Logger.Error(msgText)

			msg := tgbotapi.NewMessage(int64(AllowedChat), msgText)
			bot.Send(msg)

			continue
		}

		for _, message := range result.Messages {
			svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(sqsParams.QueueUrl),
				ReceiptHandle: message.ReceiptHandle,
			})

			var messageData MessageData
			json.Unmarshal([]byte(*message.Body), &messageData)

			if messageData.Type != "new_messages" {
				Logger.Debug("Unknown message type:", messageData.Type)
				continue
			}

			if len(messageData.Data) == 0 {
				Logger.Error("Error when parsing", *message.Body)
				continue
			}

			for _, doc := range messageData.Data {
				if len(doc.MessageType) == 0 {
					Logger.Error("Error when parsing", *message.Body)
					continue
				}

				messageHtmlWriter := bytes.Buffer{}
				err := messageTemplate.ExecuteTemplate(&messageHtmlWriter, "T", doc)
				if err != nil {
					Logger.Error("Error when rendering template:", err)
					continue
				}

				msg := tgbotapi.NewMessage(AllowedChat, messageHtmlWriter.String())
				msg.ParseMode = tgbotapi.ModeHTML
				_, sendErr := bot.Send(msg)

				if sendErr != nil {
					Logger.Error("Send message error:", sendErr)
				}
			}
		}
	}
}
