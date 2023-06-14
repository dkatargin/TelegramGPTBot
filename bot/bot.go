package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
	"telegramgptbot/gpt"
)

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
			log.Printf("[%s - %d] make request", update.Message.From.UserName, update.Message.From.ID)
			// do not send to OpenAI requests with chat commands (starst with /) and non-allowed users
			if !settings.isAllowedUser(update.Message.From.ID) || strings.HasPrefix("/", update.Message.Text) {
				log.Println("Skip request as invalid")
				continue
			}
			response, err := settings.GPTClient.Send(update.Message.Text)
			if err != nil {
				response = fmt.Sprint(err)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
			msg.ReplyToMessageID = update.Message.MessageID

			_, err = bot.Send(msg)
			if err != nil {
				logrus.WithError(err).Error("can't send answer")
			}

		}
	}
	return nil

}
