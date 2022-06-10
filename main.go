package main

import (
	"log"
	"strings"
	"wallet_tgbot/command"
	"wallet_tgbot/tg"
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

	bot, updates, err := tg.InitTgBot()
	if err != nil {
		log.Fatalf("Cannot create tgbot app:\n\t%s\n", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	paramsChan, msgChan := command.NewCommunication()

	go func() {
		for msg := range msgChan {
			m := tgbotapi.NewMessage(msg.ChatId, msg.Msg)
			m.ParseMode = "markdown"
			bot.Send(m)
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
				case "sub":
					command.SubCommand(msgChatId)
				case "del":
					command.DelCommand(msgChatId)
				}

			} else {
				params := strings.ToLower(update.Message.Text)
				paramsChan <- command.Params{ChatId: msgChatId, Msg: params}
			}
		}
	}
}
