package main

import (
	"log"
	"strings"
	"wallet_tgbot/command"
	"wallet_tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {

	err := utils.InitLogger("./logs/test.log")
	if err != nil {
		log.Fatalf("Cannot create logger:\n\t%s\n", err)
	}

	err = utils.InitEnvVar()
	if err != nil {
		log.Fatalf("Error loading .env file:\n\t%s\n", err)
	}

	bot, err := tgbotapi.NewBotAPI(utils.TgToken)
	if err != nil {
		log.Fatalf("Cannot create tgbot app:\n\t%s\n", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	paramsChan, msgChan := command.NewCommunication()

	go func() {
		for msg := range msgChan {
			bot.Send(tgbotapi.NewMessage(msg.ChatId, msg.Msg))
		}
	}()

	for update := range updates {
		if update.Message != nil {
			msgChatId := update.Message.Chat.ID

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "show":
					command.ShowCommand(msgChatId)
				case "add":
					command.AddCommand(msgChatId)
				}

			} else {
				params := strings.ToLower(update.Message.Text)
				paramsChan <- command.Params{ChatId: msgChatId, Msg: params}
			}
		}
	}
}
