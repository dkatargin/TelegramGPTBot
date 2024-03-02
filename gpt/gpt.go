package gpt

import (
	"context"
	"encoding/base64"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

type ApiClient struct {
	Token  string
	gptApi *openai.Client
}

var userContext = make(map[int64][]openai.ChatCompletionMessage)

func pushContext(msg openai.ChatCompletionMessage, userId int64) error {
	uctx, ok := userContext[userId]
	if ok {
		uctx = append(uctx, msg)
	} else {
		uctx = []openai.ChatCompletionMessage{msg}
	}
	userContext[userId] = uctx
	return nil
}
func DropUserContext(userId int64) {
	userContext[userId] = []openai.ChatCompletionMessage{}
}

func (client *ApiClient) TextRequest(text string, userId int64) (answer string, err error) {
	if client.gptApi == nil {
		client.gptApi = openai.NewClient(client.Token)
	}
	uniqUserId := fmt.Sprintf("telegram-bot-%d", userId)
	msg := openai.ChatCompletionMessage{Role: "user", Content: text, Name: uniqUserId}
	err = pushContext(msg, userId)
	if err != nil {
		logrus.WithError(err).Error("can't store user context")
	}
	resp, err := client.gptApi.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT4Turbo0125,
			Messages: userContext[userId],
			User:     uniqUserId,
		},
	)
	err = pushContext(resp.Choices[0].Message, userId)
	if err != nil {
		logrus.WithError(err).Error("can't store assistant context")
	}
	return resp.Choices[0].Message.Content, nil
}

func (client *ApiClient) ImageRequest(text string, userId int64) (answer []byte, err error) {
	if client.gptApi == nil {
		client.gptApi = openai.NewClient(client.Token)
	}
	if err != nil {
		logrus.WithError(err).Error("can't store user context")
	}
	reqBase64 := openai.ImageRequest{
		Prompt:         text,
		Size:           openai.CreateImageQualityHD,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}
	respBase64, err := client.gptApi.CreateImage(ctx, reqBase64)
	if err != nil {
		logrus.WithError(err).Error("Image creation error")
		return
	}
	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		logrus.WithError(err).Error("Base64 decode error: %v\n", err)
		return
	}
	if err != nil {
		logrus.WithError(err).Error("can't store assistant context")
	}
	return imgBytes, nil
}
