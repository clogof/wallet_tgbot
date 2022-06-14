package tg

import (
	"wallet_tgbot/command"
	"wallet_tgbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func InitTgBot() (bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, err error) {
	bot, err = tgbotapi.NewBotAPI(utils.TgToken)
	if err != nil {
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates = bot.GetUpdatesChan(u)

	return
}

func CreateMessage(msg command.ChatData) tgbotapi.MessageConfig {
	m := tgbotapi.NewMessage(msg.ChatId, msg.Msg)
	m.ParseMode = "markdown"

	var numericKeyboard tgbotapi.InlineKeyboardMarkup
	if len(msg.Params) > 0 {
		numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
		columnCount := 1

		for i, v := range msg.Params {
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
	}

	m.ReplyMarkup = numericKeyboard

	return m
}
