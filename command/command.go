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

var paramsCommandChan chan ChatData
var messageChan chan ChatData

type ChatData struct {
	ChatId int64
	Msg    string
	Params []string
}

func NewCommunication() (chan ChatData, chan ChatData) {
	paramsCommandChan = make(chan ChatData)
	messageChan = make(chan ChatData)
	modeChat = make(map[int64]int)

	go func(p chan ChatData) {
		for d := range p {
			switch modeChat[d.ChatId] {
			case add:
				addGetParams(d)
			case sub:
				subGetParams(d)
			case del:
				delGetParams(d)
			default:
				messageChan <- ChatData{ChatId: d.ChatId, Msg: "Некорректная строка"}
			}
		}
	}(paramsCommandChan)

	return paramsCommandChan, messageChan
}

func ShowCommand(chatId int64) {
	modeChat[chatId] = show

	messageChan <- ChatData{ChatId: chatId, Msg: "Получаем актуальные данные курса валют"}

	w := model.NewWallet(chatId)
	msg, err := w.Show()
	if err != nil {
		messageChan <- ChatData{ChatId: chatId, Msg: "Внутренняя ошибка"}
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Show",
			"chat_id", chatId,
			"err", err,
		)
		return
	}
	messageChan <- ChatData{ChatId: chatId, Msg: msg}
	delete(modeChat, chatId)
}

func AddCommand(chatId int64) {
	modeChat[chatId] = add

	w := model.NewWallet(chatId)
	p, err := w.GetCurrency()
	if err != nil {
		utils.Loggers.Errorw(
			"не удалось получить валюты пользователя",
			"chat_id", chatId,
		)
		messageChan <- ChatData{ChatId: chatId, Msg: "Не удалось получить валюты"}
		return
	}

	messageChan <- ChatData{
		ChatId: chatId, Msg: "Введите валюту и сумму\nНапример: btc 4.3", Params: p}
}

func addGetParams(p ChatData) {
	chatId := p.ChatId
	args := strings.Split(p.Msg, " ")

	if len(args) != 2 {
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", p.Msg,
			"chat_id", chatId,
		)
		messageChan <- ChatData{ChatId: chatId, Msg: "Некорректная строка"}
		return
	}

	coin := args[0]
	sum, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		messageChan <- ChatData{ChatId: chatId, Msg: "Некорректная сумма"}
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
		messageChan <- ChatData{ChatId: chatId, Msg: "Внутренняя ошибка"}
		return
	}

	messageChan <- ChatData{ChatId: chatId, Msg: fmt.Sprintf("Баланс %s: %f", strings.ToUpper(coin), balance)}
	delete(modeChat, p.ChatId)
}

func SubCommand(chatId int64) {
	modeChat[chatId] = sub
	messageChan <- ChatData{ChatId: chatId, Msg: "Введите валюту и сумму\nНапример: btc 4.3"}
}

func subGetParams(p ChatData) {
	chatId := p.ChatId
	args := strings.Split(p.Msg, " ")

	if len(args) != 2 {
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", strings.Join(args, " "),
			"chat_id", chatId,
		)
		messageChan <- ChatData{ChatId: chatId, Msg: "Некорректная строка"}
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
		messageChan <- ChatData{ChatId: chatId, Msg: "Некорректная сумма"}
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
		messageChan <- ChatData{ChatId: chatId, Msg: model.ErrValLessZero.Error()}
		return
	} else if err != nil {
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Sub",
			"chat_id", chatId,
			"coin", coin,
			"sum", sum,
			"err", err,
		)
		messageChan <- ChatData{ChatId: chatId, Msg: "Внутренняя ошибка"}
		return
	}
	messageChan <- ChatData{ChatId: chatId, Msg: fmt.Sprintf("Баланс %s: %f", strings.ToUpper(coin), balance)}
	delete(modeChat, p.ChatId)
}

func DelCommand(chatId int64) {
	modeChat[chatId] = del
	messageChan <- ChatData{ChatId: chatId, Msg: "Введите валюту\nНапример: btc"}
}

func delGetParams(p ChatData) {
	chatId := p.ChatId
	args := strings.Split(p.Msg, " ")

	if len(args) != 1 {
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", strings.Join(args, " "),
			"chat_id", chatId,
		)
		messageChan <- ChatData{ChatId: chatId, Msg: "Некорректная строка"}
		return
	}

	w := model.NewWallet(chatId)
	err := w.Delete(args[0])
	if errors.Is(err, model.ErrNoRowsToDel) {
		utils.Loggers.Infow(
			"Удаление валюты, отсутствующей в кошельке",
			"err", err,
		)
		messageChan <- ChatData{ChatId: chatId, Msg: model.ErrNoRowsToDel.Error()}
		return
	} else if err != nil {
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Delete",
			"chat_id", chatId,
			"coin", args[1],
			"err", err,
		)
		messageChan <- ChatData{ChatId: chatId, Msg: "Внутренняя ошибка"}
		return
	}

	messageChan <- ChatData{ChatId: chatId, Msg: "Валюта удалена"}
	delete(modeChat, p.ChatId)
}
