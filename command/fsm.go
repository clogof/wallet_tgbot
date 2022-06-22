package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"wallet_tgbot/model"
	"wallet_tgbot/utils"

	"github.com/looplab/fsm"
)

type User struct {
	ChatID     int64
	FromClient FromClientMessage
	ToClient   ToClientMessage
	State      *fsm.FSM
}

type FromClientMessage struct {
	Message string
	Args    []string
}

type ToClientMessage struct {
	Message string
	Args    []string
}

const (
	AddCommand   = "add"
	SubCommand   = "sub"
	CoinCommand  = "coin"
	ValCommand   = "val"
	StartCommand = "start"
)

const (
	ToAdd   = "toAdd"
	ToSub   = "toSub"
	ToCoin  = "toCoin"
	ToVal   = "toVal"
	ToStart = "toStart"
)

var toClientChan chan *User

func NewCommunication() (chan *User, chan *User) {
	fromClientChan := make(chan *User)
	toClientChan = make(chan *User)

	go func(fromCl chan *User) {
		var err error
		for u := range fromCl {
			switch u.State.Current() {
			case StartCommand:
				u.ToClient = ToClientMessage{Message: "Воспользуйтесь меню для выбора команды"}
				toClientChan <- u
			case AddCommand:
				err = u.Add()
				if err != nil {
					break
				}
				u.State.Event(ToCoin)
			case SubCommand:
				err = u.Sub()
				if err != nil {
					break
				}
				u.State.Event(ToCoin)
			case CoinCommand:
				err = u.Coin()
				if err != nil {
					break
				}
				u.State.Event(ToVal)
			case ValCommand:
				err = u.Val()
				if err != nil {
					break
				}
				u.State.Event(ToStart)
			default:
			}
		}
	}(fromClientChan)

	return fromClientChan, toClientChan
}

func NewUser(chatId int64) *User {
	u := &User{ChatID: chatId}

	u.State = fsm.NewFSM(
		StartCommand,
		fsm.Events{
			{Name: ToAdd, Src: []string{AddCommand, CoinCommand, SubCommand, ValCommand, StartCommand}, Dst: AddCommand},
			{Name: ToSub, Src: []string{AddCommand, CoinCommand, SubCommand, ValCommand, StartCommand}, Dst: SubCommand},
			{Name: ToCoin, Src: []string{AddCommand, SubCommand}, Dst: CoinCommand},
			{Name: ToVal, Src: []string{CoinCommand}, Dst: ValCommand},
			{Name: ToStart, Src: []string{AddCommand, CoinCommand, ValCommand}, Dst: StartCommand},
		},
		fsm.Callbacks{},
	)
	return u
}

func (u *User) Add() error {
	w := model.NewWallet(u.ChatID)

	p, err := w.GetCurrency()
	if err != nil {
		utils.Loggers.Errorw(
			"не удалось получить валюты пользователя",
			"err", err,
			"chatID", u.ChatID,
		)
		u.ToClient = ToClientMessage{Message: "Не удалось получить валюты\nПопробуйте позже"}
		toClientChan <- u
		return err
	}

	m := "Выберите валюту из списка,\nлибо введите имя новой"
	u.ToClient = ToClientMessage{Message: m, Args: p}
	u.State.SetMetadata("prev_state", "add")
	toClientChan <- u
	return nil
}

func (u *User) Sub() error {
	w := model.NewWallet(u.ChatID)

	p, err := w.GetCurrency()
	if err != nil {
		utils.Loggers.Errorw(
			"не удалось получить валюты пользователя",
			"err", err,
			"chatID", u.ChatID,
		)
		u.ToClient = ToClientMessage{Message: "Не удалось получить валюты\nПопробуйте позже"}
		toClientChan <- u
		return err
	}

	m := "Выберите валюту из списка"
	u.ToClient = ToClientMessage{Message: m, Args: p}
	u.State.SetMetadata("prev_state", "sub")
	toClientChan <- u

	return nil
}

func (u *User) Coin() error {
	u.State.SetMetadata("coin", strings.ToUpper(u.FromClient.Message))

	p, _ := u.State.Metadata("prev_state")
	c, _ := u.State.Metadata("callback")
	if p.(string) == "sub" && !c.(bool) {
		u.ToClient = ToClientMessage{Message: "Выберите значение из списка выше"}
		toClientChan <- u
		return errors.New("")
	}

	u.ToClient = ToClientMessage{Message: "Введите добавляемое/отнимаемое значение"}
	toClientChan <- u
	return nil
}

func (u *User) Val() error {
	sum, err := strconv.ParseFloat(u.FromClient.Message, 64)
	if err != nil {
		utils.Loggers.Errorw(
			"некорректное значение суммы",
			"err", err,
			"val", u.FromClient.Message,
		)
		u.ToClient = ToClientMessage{Message: "Некорректное значение суммы\nПопробуйте снова"}
		toClientChan <- u
		return err
	}

	w := model.NewWallet(u.ChatID)

	coin, _ := u.State.Metadata("coin")
	coinS := coin.(string)

	prevState, _ := u.State.Metadata("prev_state")
	prevStateS := prevState.(string)

	var balance float64
	if prevStateS == "add" {
		balance, err = w.Add(coinS, sum)
	} else if prevStateS == "sub" {
		balance, err = w.Sub(coinS, sum)
	}

	if errors.Is(err, model.ErrValLessZero) {
		utils.Loggers.Errorw(
			"вычитаемое значение больше суммы в кошелке",
			"err", err,
		)
		m := "Вычитаемое значение больше суммы в кошелке\nВведи сумму меньше"
		u.ToClient = ToClientMessage{Message: m}
		toClientChan <- u
		return err
	} else if err != nil {
		utils.Loggers.Errorw(
			"внутренняя ошибка метода",
			"err", err,
			"chatID", u.ChatID,
			"coin", coin,
			"sum", sum,
		)
		u.ToClient = ToClientMessage{Message: "Внутрення ошибка\nПопробуйте позже"}
		toClientChan <- u
		return err
	}

	u.State.SetMetadata("coin", "")
	u.State.SetMetadata("prev_state", "")

	u.ToClient = ToClientMessage{Message: fmt.Sprintf("Баланс %s: %f", coinS, balance)}
	toClientChan <- u
	return nil
}
