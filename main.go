package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gpt3 "github.com/sashabaranov/go-openai"
)

var (
	IsDebugMode   bool
	TelegramToken string
	BotMembers    []int64
	ChatGPTToken  string

	gptClient *gpt3.Client
	ctx       = context.Background()
)

func isAllowedUser(userId int64) bool {
	for _, i := range BotMembers {
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

func initConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		logrus.WithError(err).Fatalf("can't read config: %s\n", configPath)
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		pair := strings.SplitN(line, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := strings.TrimSpace(pair[0])
		val := strings.TrimSpace(pair[1])
		if key == "telegram_token" {
			TelegramToken = val
		} else if key == "chatgpt_token" {
			ChatGPTToken = val
		} else if key == "telegram_admins" {
			admins := strings.Split(val, ",")
			adminList := make([]int64, 0)
			for _, strUid := range admins {
				intUid, err := strconv.ParseInt(strUid, 10, 64)
				if err != nil {
					logrus.WithError(err).Warnf("possible wrong telegram id: %s\n", strUid)
					continue
				}
				adminList = append(adminList, intUid)
			}
			BotMembers = adminList
		} else if key == "mode" {
			IsDebugMode = val == "debug"
		}
	}

	return nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.cfg", "path to config file")
	err := initConfig(configPath)
	if err != nil {
		panic(err)
	}
	if IsDebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	logrus.Infof("config successfully init with admins %v", BotMembers)
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

			_, err = bot.Send(msg)
			if err != nil {
				logrus.WithError(err).Error("can't send answer")
			}

		}
	}

}
