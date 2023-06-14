package main

import (
	"bufio"
	"flag"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"telegramgptbot/bot"
	"telegramgptbot/gpt"
)

var (
	IsDebugMode   bool
	TelegramToken string
	BotMembers    []int64
	ChatGPTClient *gpt.ApiClient
	TelegramBot   *bot.TelegramBot
)

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
		} else if key == "openai_token" {
			ChatGPTClient = &gpt.ApiClient{
				Token: val,
			}
		} else if key == "bot_members" {
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
	// configure bot
	TelegramBot = &bot.TelegramBot{
		Token:      TelegramToken,
		BotMembers: BotMembers,
		GPTClient:  ChatGPTClient,
	}
	err = TelegramBot.Handle()
	if err != nil {
		panic(err)
	}

}
