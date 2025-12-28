package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

const AllowedUserName = "andre487"
const AllowedChat = 94764326

var lockBoxHandler = "https://payload.lockbox.api.cloud.yandex.net/lockbox/v1/secrets"
var secretDir string

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
	Code      int    `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}

type SecretProvider interface {
	Init()
	BotToken() string
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

func (m DevSecretProvider) SqsParams() SqsParamsData {
	return SqsParamsData{
		// QueueUrl:  readSecretFile("sqs-prod-queue"),
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

func (m YcSecretProvider) SqsParams() SqsParamsData {
	return SqsParamsData{
		QueueUrl:  requestLockBoxTextValue("e6qq93te4b88t6qv2ak0", "prod-queue"),
		AccessKey: requestLockBoxTextValue("e6qq93te4b88t6qv2ak0", "access-key"),
		SecretKey: requestLockBoxTextValue("e6qq93te4b88t6qv2ak0", "secret-key"),
	}
}

func readSecretFile(filePath string) string {
	res, err := os.ReadFile(path.Join(secretDir, filePath))
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
	if len(iamToken) == 0 {
		Logger.Fatal("Empty IAM token")
	}

	url := fmt.Sprintf("%s/%s/payload", lockBoxHandler, secId)

	req, err := http.NewRequest("GET", url, nil)
	PanicOnErr(err)
	req.Header.Set("Authorization", "Bearer "+iamToken)

	client := http.Client{}
	res, err := client.Do(req)
	PanicOnErr(err)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			Logger.Warning("Error closing response body:", err)
		}
	}()

	resultBytes, err := io.ReadAll(res.Body)
	PanicOnErr(err)

	var result LockBoxResult
	err = json.Unmarshal(resultBytes, &result)
	PanicOnErr(err)
	if result.Code != 0 {
		Logger.Fatal(result.Message)
	}

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
	defer func() {
		err := res.Body.Close()
		if err != nil {
			Logger.Warning("Error closing response body:", err)
		}
	}()

	resultBytes, err := io.ReadAll(res.Body)
	PanicOnErr(err)

	var tokenData IamTokenData
	err = json.Unmarshal(resultBytes, &tokenData)
	PanicOnErr(err)

	if len(tokenData.AccessToken) == 0 {
		return "", errors.New("no IAM token")
	}
	return tokenData.AccessToken, nil
}

func requestLockBoxTextValue(secId string, name string) string {
	result := requestLockBox(secId)
	value := ""
	matched := false
	for _, val := range result.Entries {
		if val.Key == name {
			value = val.TextValue
			matched = true
			break
		}
	}
	if !matched {
		Logger.Fatalf("%s not found in secret '%s'", name, secId)
	}
	if len(value) == 0 {
		Logger.Warning(name + " is empty")
	}
	return value
}
