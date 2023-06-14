package gpt

import (
	"context"
	"github.com/sashabaranov/go-openai"
	gpt3 "github.com/sashabaranov/go-openai"
)

var ctx = context.Background()

type ApiClient struct {
	Token  string
	gptApi *openai.Client
}

func (client *ApiClient) Send(text string, contextMessages []string) (answer string, err error) {
	if client.gptApi == nil {
		client.gptApi = gpt3.NewClient(client.Token)
	}
	resp, err := client.gptApi.CreateChatCompletion(
		ctx,
		gpt3.ChatCompletionRequest{
			Model: gpt3.GPT3Dot5Turbo,
			Messages: []gpt3.ChatCompletionMessage{
				{Role: "user", Content: text},
			},
		},
	)
	return resp.Choices[0].Message.Content, nil
}
