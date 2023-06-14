package gpt

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	gpt3 "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"log"
)

var ctx = context.Background()

type ApiClient struct {
	Token  string
	gptApi *openai.Client
}

var userContext = make(map[int64][]gpt3.ChatCompletionMessage)

func pushContext(msg gpt3.ChatCompletionMessage, userId int64) error {
	uctx, ok := userContext[userId]
	if ok {
		uctx = append(uctx, msg)
	} else {
		uctx = []gpt3.ChatCompletionMessage{msg}
	}
	userContext[userId] = uctx
	return nil
}
func DropUserContext(userId int64) {
	userContext[userId] = []gpt3.ChatCompletionMessage{}
}

func (client *ApiClient) Send(text string, userId int64) (answer string, err error) {
	if client.gptApi == nil {
		client.gptApi = gpt3.NewClient(client.Token)
	}
	uniqUserId := fmt.Sprintf("telegram-bot-%d", userId)
	msg := gpt3.ChatCompletionMessage{Role: "user", Content: text, Name: uniqUserId}
	err = pushContext(msg, userId)
	if err != nil {
		logrus.WithError(err).Error("can't store user context")
	}
	log.Println(userContext[userId])
	resp, err := client.gptApi.CreateChatCompletion(
		ctx,
		gpt3.ChatCompletionRequest{
			Model:    gpt3.GPT3Dot5Turbo,
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
