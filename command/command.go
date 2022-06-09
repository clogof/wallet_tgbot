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
}

func SubCommand(args []string, chat_id int64) (msg string) {
	if len(args) != 3 {
		msg = "Некорректная строка"
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", strings.Join(args, " "),
			"chat_id", chat_id,
		)
		return
	}

	coin := args[1]
	sum, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		msg = "некорректная сумма"
		utils.Loggers.Errorw(
			"Некорректное значение суммы",
			"val", args[2],
			"err", err,
		)
		return
	}

	w := model.NewWallet(chat_id)
	balance, err := w.Sub(coin, sum)

	if errors.Is(err, model.ErrValLessZero) {
		msg = model.ErrValLessZero.Error()
		utils.Loggers.Infow(
			"вычитаемое значение больше суммы в кошелке",
			"sub_coin", coin,
			"err", err,
		)
		return
	} else if err != nil {
		msg = "Внутренняя ошибка"
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Sub",
			"chat_id", chat_id,
			"coin", coin,
			"sum", sum,
			"err", err,
		)
		return
	}
	msg = fmt.Sprintf("Баланс %s: %f", strings.ToUpper(coin), balance)
	return
}

func DelCommand(args []string, chat_id int64) (msg string) {
	if len(args) != 2 {
		msg = "Некорректная строка"
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", strings.Join(args, " "),
			"chat_id", chat_id,
		)
		return
	}

	w := model.NewWallet(chat_id)
	err := w.Delete(args[1])
	if errors.Is(err, model.ErrNoRowsToDel) {
		msg = model.ErrNoRowsToDel.Error()
		utils.Loggers.Infow(
			"Удаление валюты, отсутствующей в кошельке",
			"err", err,
		)
		return
	} else if err != nil {
		msg = "Внутренняя ошибка"
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Delete",
			"chat_id", chat_id,
			"coin", args[1],
			"err", err,
		)
		return
	}

	msg = "Валюта удалена"
	return
}
