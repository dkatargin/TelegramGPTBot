package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gpt3 "github.com/sashabaranov/go-openai"
)

const (
	TelegramToken = "XXXXX" // Telegram API Token from @BotFather
	ChatGPTToken  = "XXXXX" // OpenAI token from https://platform.openai.com/account/api-keys
)

var (
	gptClient *gpt3.Client
	ctx       = context.Background()
	botUsers  = []int64{00000} // List of allowed Telegram user id's
)


func isAllowedUser(userId int64) bool {
	for _, i := range botUsers {
		if userId == i {
			return true
		}
	}
	return false
}

func sendChatRequest(text *string) (answer *string, err error) {
	if gptClient == nil {
		gptClient = gpt3.NewClient(ChatGPTToken)
	}
	resp, err := gptClient.CreateChatCompletion(
		ctx,
		gpt3.ChatCompletionRequest{
			Model: gpt3.GPT3Dot5Turbo,
			Messages: []gpt3.ChatCompletionMessage{
				{Role: "user", Content: *text},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	return &resp.Choices[0].Message.Content, nil
}

func main() {
	var responseText string
	bot, err := tgbotapi.NewBotAPI(TelegramToken)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
  
  // handle telegram bot messages
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s - %d] make request", update.Message.From.UserName, update.Message.From.ID)
      // do not send to OpenAI requests with chat commands (starst with /) and non-allowed users
			if !isAllowedUser(update.Message.From.ID) || strings.HasPrefix("/", update.Message.Text) {
				log.Println("Skip request as invalid")
				continue
			}
			response, err := sendChatRequest(&update.Message.Text)
			if err != nil {
				responseText = fmt.Sprint(err)
			} else {
				responseText = *response
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)

		}
	}

}
