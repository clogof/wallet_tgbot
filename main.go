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

	var msg string
	for update := range updates {
		if update.Message != nil {

			msgText := strings.ToLower(update.Message.Text)
			msgArr := strings.Split(msgText, " ")

			switch msgArr[0] {
			case "add":
				msg = command.AddCommand(msgArr, update.Message.Chat.ID)
			case "sub":
				msg = command.SubCommand(msgArr, update.Message.Chat.ID)
			case "del":
				msg = command.DelCommand(msgArr, update.Message.Chat.ID)
			case "show":
				msg = command.ShowCommand(update.Message.Chat.ID)
			default:
				msg = "Некорректная команда"
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		}
	}
}
