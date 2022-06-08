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
	show = iota
	add
	sub
	del
)

var Mode int

var ParamsCommandChan chan Params
var Message chan string

type Params struct {
	ChatId int64
	Args   []string
}

func AddCommand(args []string, chat_id int64) (msg string) {
	Mode = add
	msg = "Введите валюту и сумму"
	return
}

func addCom(p Params) {
	if len(p.Args) != 2 {
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", strings.Join(p.Args, " "),
			"chat_id", p.ChatId,
		)
		Message <- "Некорректная строка"
		return
	}

	coin := p.Args[0]
	sum, err := strconv.ParseFloat(p.Args[1], 64)
	if err != nil {
		Message <- "Некорректная сумма"
		utils.Loggers.Errorw(
			"некорректное значение суммы",
			"val", p.Args[1],
			"err", err,
		)
		return
	}

	w := model.NewWallet(p.ChatId)
	balance, err := w.Add(coin, sum)
	if err != nil {
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Add",
			"chat_id", p.ChatId,
			"coin", coin,
			"sum", sum,
			"err", err,
		)
		Message <- "Внутренняя ошибка"
		return
	}

	Message <- fmt.Sprintf("Баланс %s: %f", strings.ToUpper(coin), balance)
}

func H() {
	ParamsCommandChan = make(chan Params)
	Message = make(chan string)

	for {
		d := <-ParamsCommandChan
		switch Mode {
		case add:
			addCom(d)
		}

	}
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

func ShowCommand(chat_id int64) (msg string) {
	Mode = show

	w := model.NewWallet(chat_id)
	msg, err := w.Show()
	if err != nil {
		msg = "Внутренняя ошибка"
		utils.Loggers.Errorw(
			"внутренняя ошибка метода Show",
			"chat_id", chat_id,
			"err", err,
		)
		return
	}
	return
}
