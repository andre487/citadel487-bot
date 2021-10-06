package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var secretDir string

type S3ParamsData struct {
	AccessKey string
	SecretKey string
}

type SqsParamsData struct {
	QueueUrl  string
	AccessKey string
	SecretKey string
}

type SecretProvider interface {
	Init()
	BotToken() string
	Netrc() string
	S3Params() S3ParamsData
	SqsParams() SqsParamsData
}

type DevSecretProvider struct {
}

func NewSecretProvider() SecretProvider {
	sp := DevSecretProvider{}
	sp.Init()
	return sp
}

func (m DevSecretProvider) Init() {
	cwd, err := os.Getwd()
	PanicOnErr(err)
	secretDir = path.Join(cwd, ".secrets")
}

func (m DevSecretProvider) BotToken() string {
	return m.readSecretFile("dev-bot-token")
}

func (m DevSecretProvider) Netrc() string {
	return m.readSecretFile("netrc")
}

func (m DevSecretProvider) S3Params() S3ParamsData {
	return S3ParamsData{
		AccessKey: m.readSecretFile("s3-access"),
		SecretKey: m.readSecretFile("s3-secret"),
	}
}

func (m DevSecretProvider) SqsParams() SqsParamsData {
	return SqsParamsData{
		QueueUrl:  m.readSecretFile("sqs-test-queue"),
		AccessKey: m.readSecretFile("sqs-access-key"),
		SecretKey: m.readSecretFile("sqs-secret-key"),
	}
}

func (m DevSecretProvider) readSecretFile(filePath string) string {
	res, err := ioutil.ReadFile(path.Join(secretDir, filePath))
	PanicOnErr(err)
	return strings.TrimSpace(string(res))
}
