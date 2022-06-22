package command

import (
	"fmt"
	"strconv"
	"wallet_tgbot/model"

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
			case AddCommand:
				err = u.Add()
				if err != nil {
					fmt.Printf("err: %s\n", err)
				}
				u.State.Event(ToCoin)
			case CoinCommand:
				err = u.Coin()
				if err != nil {
					fmt.Printf("err: %s\n", err)
				}
				u.State.Event(ToVal)
			case ValCommand:
				err = u.Val()
				if err != nil {
					fmt.Printf("err: %s\n", err)
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
	p, _ := w.GetCurrency()
	// logger
	u.ToClient = ToClientMessage{Message: "Выберите валюту", Args: p}
	toClientChan <- u
	return nil
}

func (u *User) Coin() error {
	u.State.SetMetadata("coin", u.FromClient.Message)
	u.ToClient = ToClientMessage{Message: "Введите значение"}
	toClientChan <- u
	return nil
}

func (u *User) Val() error {
	sum, _ := strconv.ParseFloat(u.FromClient.Message, 64)
	// logger
	w := model.NewWallet(u.ChatID)
	coin, _ := u.State.Metadata("coin")
	u.State.SetMetadata("coin", "")

	r, _ := w.Add(coin.(string), sum)
	// logger
	u.ToClient = ToClientMessage{Message: fmt.Sprintf("Новое значение: %f", r)}
	toClientChan <- u
	return nil
}
