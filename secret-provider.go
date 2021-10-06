package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

var lockBoxHandler = "https://payload.lockbox.api.cloud.yandex.net/lockbox/v1/secrets"
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

type IamTokenData struct {
	AccessToken string `json:"access_token"`
}

type LockBoxResult struct {
	Entries []struct {
		Key         string
		TextValue   string
		BinaryValue string
	}
	VersionId string
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

type YcSecretProvider struct {
}

func NewSecretProvider() SecretProvider {
	var sp SecretProvider
	deployType := os.Getenv("DEPLOY_TYPE")
	Logger.Info("Deploy type:", deployType)
	if deployType == "prod" {
		sp = YcSecretProvider{}
	} else {
		sp = DevSecretProvider{}
	}
	sp.Init()
	return sp
}

func (m DevSecretProvider) Init() {
	cwd, err := os.Getwd()
	PanicOnErr(err)
	secretDir = path.Join(cwd, ".secrets")
}

func (m DevSecretProvider) BotToken() string {
	return readSecretFile("dev-bot-token")
}

func (m DevSecretProvider) Netrc() string {
	return readSecretFile("netrc")
}

func (m DevSecretProvider) S3Params() S3ParamsData {
	return S3ParamsData{
		AccessKey: readSecretFile("s3-access"),
		SecretKey: readSecretFile("s3-secret"),
	}
}

func (m DevSecretProvider) SqsParams() SqsParamsData {
	return SqsParamsData{
		QueueUrl:  readSecretFile("sqs-test-queue"),
		AccessKey: readSecretFile("sqs-access-key"),
		SecretKey: readSecretFile("sqs-secret-key"),
	}
}

func (m YcSecretProvider) Init() {
}

func (m YcSecretProvider) BotToken() string {
	return requestLockBoxTextValue("e6qbv0lnihrdt4mmer19", "token")
}

func (m YcSecretProvider) Netrc() string {
	return requestLockBoxBinaryValue("e6qmn9f60sspf916ncu1", "content")
}

func (m YcSecretProvider) S3Params() S3ParamsData {
	return S3ParamsData{
		AccessKey: requestLockBoxTextValue("e6q3nf38hdbee440d4l8", "s3-access"),
		SecretKey: requestLockBoxTextValue("e6q3nf38hdbee440d4l8", "s3-secret"),
	}
}

func (m YcSecretProvider) SqsParams() SqsParamsData {
	return SqsParamsData{
		QueueUrl:  requestLockBoxTextValue("e6qq93te4b88t6qv2ak0", "prod-queue"),
		AccessKey: requestLockBoxTextValue("e6qq93te4b88t6qv2ak0", "access-key"),
		SecretKey: requestLockBoxTextValue("e6qq93te4b88t6qv2ak0", "secret-key"),
	}
}

func readSecretFile(filePath string) string {
	res, err := ioutil.ReadFile(path.Join(secretDir, filePath))
	PanicOnErr(err)
	return strings.TrimSpace(string(res))
}

func requestLockBox(secId string) LockBoxResult {
	iamToken, err := getIamToken()
	if err != nil {
		Logger.Error("Error when getting IAM token:", err)

		cmd := exec.Command("yc", "iam", "create-token")
		var tokenBuffer bytes.Buffer
		cmd.Stdout = &tokenBuffer
		err := cmd.Run()
		PanicOnErr(err)

		iamToken = strings.TrimSpace(tokenBuffer.String())
	}

	url := fmt.Sprintf("%s/%s/payload", lockBoxHandler, secId)

	req, err := http.NewRequest("GET", url, nil)
	PanicOnErr(err)
	req.Header.Set("Authorization", "Bearer "+iamToken)

	client := http.Client{}
	res, err := client.Do(req)
	PanicOnErr(err)

	resultBytes, err := ioutil.ReadAll(res.Body)
	PanicOnErr(err)

	var result LockBoxResult
	json.Unmarshal(resultBytes, &result)
	return result
}

func getIamToken() (string, error) {
	metaServiceHost := os.Getenv("YC_METADATA_SERVICE")
	if len(metaServiceHost) == 0 {
		metaServiceHost = "169.254.169.254"
	}
	url := fmt.Sprintf("http://%s/computeMetadata/v1/instance/service-accounts/default/token", metaServiceHost)

	req, err := http.NewRequest("GET", url, nil)
	PanicOnErr(err)
	req.Header.Set("Metadata-Flavor", "Google")

	client := http.Client{}
	res, err := client.Do(req)
	PanicOnErr(err)

	resultBytes, err := ioutil.ReadAll(res.Body)
	PanicOnErr(err)

	var tokenData IamTokenData
	json.Unmarshal(resultBytes, &tokenData)

	if len(tokenData.AccessToken) == 0 {
		return "", errors.New("no IAM token")
	}
	return tokenData.AccessToken, nil
}

func requestLockBoxTextValue(secId string, name string) string {
	result := requestLockBox(secId)
	value := ""
	for _, val := range result.Entries {
		if val.Key == name {
			value = val.TextValue
			break
		}
	}
	if len(value) == 0 {
		Logger.Warning(name + " is empty")
	}
	return value
}

func requestLockBoxBinaryValue(secId string, name string) string {
	result := requestLockBox(secId)
	value := ""
	for _, val := range result.Entries {
		if val.Key == name {
			value = val.BinaryValue
			break
		}
	}
	if len(value) == 0 {
		Logger.Warning(name + " is empty")
		return ""
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	PanicOnErr(err)

	return string(decoded)
}
