package tg

import (
	"log"
	"wallet_tgbot/command"
	"wallet_tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI

func InitTgBot() (tgbotapi.UpdatesChannel, error) {
	var err error
	bot, err = tgbotapi.NewBotAPI(utils.TgToken)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return updates, nil
}

func SendMessage(msg *command.User) {
	m := tgbotapi.NewMessage(msg.ChatID, msg.ToClient.Message)
	m.ParseMode = "markdown"

	var numericKeyboard tgbotapi.InlineKeyboardMarkup
	if len(msg.ToClient.Args) > 0 {
		numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
		columnCount := 1

		for i, v := range msg.ToClient.Args {
			if i%columnCount == 0 {
				numericKeyboard.InlineKeyboard = append(
					numericKeyboard.InlineKeyboard,
					tgbotapi.NewInlineKeyboardRow(),
				)
			}

			numericKeyboard.InlineKeyboard[i/columnCount] = append(
				numericKeyboard.InlineKeyboard[i/columnCount],
				tgbotapi.NewInlineKeyboardButtonData(v, v),
			)
		}
		m.ReplyMarkup = numericKeyboard
	}

	bot.Send(m)
}

func CheckCallback(msg *command.User) {
	callback := tgbotapi.NewCallback(msg.FromClient.Args[0], msg.FromClient.Message)
	if _, err := bot.Request(callback); err != nil {
		log.Printf("CheckCallback error: %s", err)
	}
	bot.Send(tgbotapi.NewMessage(msg.ChatID, "Выбрано "+msg.FromClient.Message))
}
