package tg

import (
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
