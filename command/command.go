package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"wallet_tgbot/model"
	"wallet_tgbot/utils"
)

const (
	show = iota + 1
	add
	sub
	del
)

var modeChat map[int64]int

var paramsCommandChan chan Params
var messageChan chan Params

type Params struct {
	ChatId int64
	Msg    string
}

func NewCommunication() (chan Params, chan Params) {
	paramsCommandChan = make(chan Params)
	messageChan = make(chan Params)
	modeChat = make(map[int64]int)

	go func(p chan Params) {
		for d := range p {
			switch modeChat[d.ChatId] {
			case add:
				addGetParams(d)
			case sub:
				subGetParams(d)
			case del:
				delGetParams(d)
			default:
				messageChan <- Params{ChatId: d.ChatId, Msg: "Некорректная строка"}
			}
		}
	}(paramsCommandChan)

	return paramsCommandChan, messageChan
}

func ShowCommand(chatId int64) {
	modeChat[chatId] = show

	messageChan <- Params{ChatId: chatId, Msg: "Получаем актуальные данные курса валют"}

	w := model.NewWallet(chatId)
	msg, err := w.Show()
	if err != nil {
		messageChan <- Params{ChatId: chatId, Msg: "Внутренняя ошибка"}
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Show",
			"chat_id", chatId,
			"err", err,
		)
		return
	}
	messageChan <- Params{ChatId: chatId, Msg: msg}
	delete(modeChat, chatId)
}

func AddCommand(chatId int64) {
	modeChat[chatId] = add
	messageChan <- Params{ChatId: chatId, Msg: "Введите валюту и сумму\nНапример: btc 4.3"}
}

func addGetParams(p Params) {
	chatId := p.ChatId
	args := strings.Split(p.Msg, " ")

	if len(args) != 2 {
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", p.Msg,
			"chat_id", chatId,
		)
		messageChan <- Params{ChatId: chatId, Msg: "Некорректная строка"}
		return
	}

	coin := args[0]
	sum, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		messageChan <- Params{ChatId: chatId, Msg: "Некорректная сумма"}
		utils.Loggers.Errorw(
			"некорректное значение суммы",
			"val", args[1],
			"err", err,
		)
		return
	}

	w := model.NewWallet(chatId)
	balance, err := w.Add(coin, sum)
	if err != nil {
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Add",
			"chat_id", chatId,
			"coin", coin,
			"sum", sum,
			"err", err,
		)
		messageChan <- Params{ChatId: chatId, Msg: "Внутренняя ошибка"}
		return
	}

	messageChan <- Params{ChatId: chatId, Msg: fmt.Sprintf("Баланс %s: %f", strings.ToUpper(coin), balance)}
	delete(modeChat, p.ChatId)
}

func SubCommand(chatId int64) {
	modeChat[chatId] = sub
	messageChan <- Params{ChatId: chatId, Msg: "Введите валюту и сумму\nНапример: btc 4.3"}
}

func subGetParams(p Params) {
	chatId := p.ChatId
	args := strings.Split(p.Msg, " ")

	if len(args) != 2 {
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", strings.Join(args, " "),
			"chat_id", chatId,
		)
		messageChan <- Params{ChatId: chatId, Msg: "Некорректная строка"}
		return
	}

	coin := args[0]
	sum, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		utils.Loggers.Errorw(
			"некорректное значение суммы",
			"val", args[1],
			"err", err,
		)
		messageChan <- Params{ChatId: chatId, Msg: "Некорректная сумма"}
		return
	}

	w := model.NewWallet(chatId)
	balance, err := w.Sub(coin, sum)

	if errors.Is(err, model.ErrValLessZero) {
		utils.Loggers.Infow(
			"вычитаемое значение больше суммы в кошелке",
			"sub_coin", coin,
			"err", err,
		)
		messageChan <- Params{ChatId: chatId, Msg: model.ErrValLessZero.Error()}
		return
	} else if err != nil {
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Sub",
			"chat_id", chatId,
			"coin", coin,
			"sum", sum,
			"err", err,
		)
		messageChan <- Params{ChatId: chatId, Msg: "Внутренняя ошибка"}
		return
	}
	messageChan <- Params{ChatId: chatId, Msg: fmt.Sprintf("Баланс %s: %f", strings.ToUpper(coin), balance)}
	delete(modeChat, p.ChatId)
}

func DelCommand(chatId int64) {
	modeChat[chatId] = del
	messageChan <- Params{ChatId: chatId, Msg: "Введите валюту\nНапример: btc"}
}

func delGetParams(p Params) {
	chatId := p.ChatId
	args := strings.Split(p.Msg, " ")

	if len(args) != 1 {
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", strings.Join(args, " "),
			"chat_id", chatId,
		)
		messageChan <- Params{ChatId: chatId, Msg: "Некорректная строка"}
		return
	}

	w := model.NewWallet(chatId)
	err := w.Delete(args[0])
	if errors.Is(err, model.ErrNoRowsToDel) {
		utils.Loggers.Infow(
			"Удаление валюты, отсутствующей в кошельке",
			"err", err,
		)
		messageChan <- Params{ChatId: chatId, Msg: model.ErrNoRowsToDel.Error()}
		return
	} else if err != nil {
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Delete",
			"chat_id", chatId,
			"coin", args[1],
			"err", err,
		)
		messageChan <- Params{ChatId: chatId, Msg: "Внутренняя ошибка"}
		return
	}

	messageChan <- Params{ChatId: chatId, Msg: "Валюта удалена"}
	delete(modeChat, p.ChatId)
}
