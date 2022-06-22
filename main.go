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

	updates, err := tg.InitTgBot()
	if err != nil {
		log.Fatalf("Cannot create tgbot app:\n\t%s\n", err)
	}

	fromClientChan, toClientChan := command.NewCommunication()
	users := make(map[int64]*command.User)

	go func() {
		for msg := range toClientChan {
			tg.SendMessage(msg)
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
				case command.AddCommand:
					users[msgChatId].FromClient = command.FromClientMessage{}
					users[msgChatId].State.Event(command.ToAdd)
					fromClientChan <- users[msgChatId]
				case command.SubCommand:
					users[msgChatId].FromClient = command.FromClientMessage{}
					users[msgChatId].State.Event(command.ToSub)
					fromClientChan <- users[msgChatId]
				case command.DelCommand:
					users[msgChatId].FromClient = command.FromClientMessage{}
					users[msgChatId].State.Event(command.ToDel)
					fromClientChan <- users[msgChatId]
				default:
					m := "Некорректна команда\nВоспользуйтесь меню"
					users[msgChatId].ToClient = command.ToClientMessage{Message: m}
					toClientChan <- users[msgChatId]
				}
			} else {
				m := strings.ToLower(update.Message.Text)
				users[msgChatId].FromClient = command.FromClientMessage{Message: m}
				users[msgChatId].State.SetMetadata("callback", false)
				fromClientChan <- users[msgChatId]
			}
		} else if update.CallbackQuery != nil {
			msgChatId := update.CallbackQuery.Message.Chat.ID
			data := update.CallbackQuery.Data

			tg.CheckCallback(&command.User{
				ChatID: msgChatId,
				FromClient: command.FromClientMessage{
					Message: data,
					Args:    []string{update.CallbackQuery.ID}}},
			)

			users[msgChatId].FromClient = command.FromClientMessage{Message: data}
			users[msgChatId].State.SetMetadata("callback", true)
			fromClientChan <- users[msgChatId]
		}
	}
}
