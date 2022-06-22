package command

import (
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
	CoinCommand  = "coin"
	ValCommand   = "val"
	StartCommand = "start"
)

const (
	ToAdd   = "toAdd"
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
			{Name: ToAdd, Src: []string{AddCommand, CoinCommand, ValCommand, StartCommand}, Dst: AddCommand},
			{Name: ToCoin, Src: []string{AddCommand}, Dst: CoinCommand},
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
	toClientChan <- u
	return nil
}

func (u *User) Coin() error {
	u.State.SetMetadata("coin", strings.ToUpper(u.FromClient.Message))
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
	u.State.SetMetadata("coin", "")

	balance, err := w.Add(coinS, sum)
	if err != nil {
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

	u.ToClient = ToClientMessage{Message: fmt.Sprintf("Баланс %s: %f", coinS, balance)}
	toClientChan <- u
	return nil
}
