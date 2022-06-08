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

	go command.H()

	i := 0

	for update := range updates {
		if update.Message != nil {
			log.Printf("i: %d", i)
			msg := ""
			msgChatId := update.Message.Chat.ID
			// if i%2 == 0 {
			// 	msgChatId = 666
			// } else {
			// 	msgChatId = 777
			// }

			log.Printf(
				"id: %d command: %s msg: %s",
				msgChatId, update.Message.Command(), update.Message.Text,
			)

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "show":
					msg = command.ShowCommand(msgChatId)
				case "add":
					msg = command.AddCommand(msgChatId)
				}

			} else {
				params := strings.ToLower(update.Message.Text)

				command.ParamsCommandChan <- params
				msg = <-command.MessageChan
			}

			bot.Send(tgbotapi.NewMessage(msgChatId, msg))
			i++
		}
	}
}
