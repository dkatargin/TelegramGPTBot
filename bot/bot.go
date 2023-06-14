package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
	"telegramgptbot/gpt"
	"unicode/utf8"
)

const TgMaxMessageLen = 4096

type TelegramBot struct {
	BotMembers []int64
	Token      string
	GPTClient  *gpt.ApiClient
}

func (settings *TelegramBot) isAllowedUser(userId int64) bool {
	for _, i := range settings.BotMembers {
		if userId == i {
			return true
		}
	}
	return false
}

func splitMessages(gptResponse string) []string {
	// TODO: split code (```) correctly
	messages := make([]string, 0)
	var l, r int
	for l, r = 0, TgMaxMessageLen; r < len(gptResponse); l, r = r, r+TgMaxMessageLen {
		for !utf8.RuneStart(gptResponse[r]) {
			r--
		}
		messages = append(messages, gptResponse[l:r])
	}
	messages = append(messages, gptResponse[l:])
	return messages
}

func (settings *TelegramBot) Handle() error {
	bot, err := tgbotapi.NewBotAPI(settings.Token)
	if err != nil {
		return err
	}
	updatesChan := tgbotapi.NewUpdate(0)
	updatesChan.Timeout = 60

	updates := bot.GetUpdatesChan(updatesChan)
	for update := range updates {
		if update.Message != nil {
			contextMessages := make([]string, 0)
			log.Printf("[%s - %d] make request", update.Message.From.UserName, update.Message.From.ID)
			// do not send to OpenAI requests with chat commands (starst with /) and non-allowed users
			if !settings.isAllowedUser(update.Message.From.ID) || strings.HasPrefix("/", update.Message.Text) {
				log.Println("Skip request as invalid")
				continue
			}
			// send request to GPT
			response, err := settings.GPTClient.Send(update.Message.Text, contextMessages)
			if err != nil {
				logrus.WithError(err).Error("can't send request")
				response = fmt.Sprint(err)
			}
			// send response to chat with user
			replyTo := update.Message.MessageID
			for _, m := range splitMessages(response) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
				msg.ReplyToMessageID = replyTo
				newMsg, err := bot.Send(msg)
				if err != nil {
					logrus.WithError(err).Error("can't send answer")
				}
				replyTo = newMsg.MessageID
			}
		}
	}
	return nil

}
