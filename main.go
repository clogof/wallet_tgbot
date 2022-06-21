package main

import (
	"log"
	"strings"
	"wallet_tgbot/command"
	"wallet_tgbot/tg"
	"wallet_tgbot/utils"
)

func main() {
	var err error

	if err = utils.InitLogger("./logs/test.log"); err != nil {
		log.Fatalf("Cannot create logger:\n\t%s\n", err)
	}

	if err = utils.InitEnvVar(); err != nil {
		log.Fatalf("Error loading .env file:\n\t%s\n", err)
	}

	bot, updates, err := tg.InitTgBot()
	if err != nil {
		log.Fatalf("Cannot create tgbot app:\n\t%s\n", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	fromClientChan, toClientChan := command.NewCommunication()
	users := make(map[int64]*command.User)

	go func() {
		for msg := range toClientChan {
			m := tg.CreateMessage(msg)
			bot.Send(m)
		}
	}()

	for update := range updates {
		if update.Message != nil {
			msgChatId := update.Message.Chat.ID

			if _, ok := users[msgChatId]; !ok {
				users[msgChatId] = command.NewUser(msgChatId)
			}

			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "add":
					users[msgChatId].FromClient = command.FromClientMessage{}
					users[msgChatId].State.Event("toAdd")
					fromClientChan <- users[msgChatId]
				default:
				}
			} else {
				m := strings.ToLower(update.Message.Text)
				users[msgChatId].FromClient = command.FromClientMessage{Message: m}
				fromClientChan <- users[msgChatId]
			}
		} else if update.CallbackQuery != nil {
			msgChatId := update.CallbackQuery.Message.Chat.ID
			data := update.CallbackQuery.Data

			users[msgChatId].FromClient = command.FromClientMessage{Message: data}
			fromClientChan <- users[msgChatId]
		}
	}
}
