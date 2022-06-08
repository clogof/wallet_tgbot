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

var mode int
var chatId int64

var ParamsCommandChan chan string
var MessageChan chan string

func H() {
	ParamsCommandChan = make(chan string)
	MessageChan = make(chan string)

	for {
		d := <-ParamsCommandChan
		switch mode {
		case add:
			addCom(d)
		}

	}
}

func AddCommand(chat_id int64) (msg string) {
	mode = add
	chatId = chat_id
	msg = "Введите валюту и сумму\nНапример: btc 4.3"
	return
}

func addCom(s string) {
	args := strings.Split(s, " ")

	if len(args) != 2 {
		utils.Loggers.Errorw(
			"некорректная строка",
			"args", s,
			"chat_id", chatId,
		)
		MessageChan <- "Некорректная строка"
		return
	}

	coin := args[0]
	sum, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		MessageChan <- "Некорректная сумма"
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
		MessageChan <- "Внутренняя ошибка"
		return
	}

	MessageChan <- fmt.Sprintf("Баланс %s: %f", strings.ToUpper(coin), balance)
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
	mode = show

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
