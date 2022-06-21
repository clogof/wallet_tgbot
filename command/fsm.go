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
	add  = "add"
	coin = "coin"
	val  = "val"
	menu = "menu"
)

var toClientChan chan *User

func NewCommunication() (chan *User, chan *User) {
	fromClientChan := make(chan *User)
	toClientChan = make(chan *User)

	go func(fromCl chan *User) {
		var err error
		for u := range fromCl {
			switch u.State.Current() {
			case add:
				err = u.Add()
				if err != nil {
					fmt.Printf("err: %s\n", err)
				}
				u.State.Event("toCoin")
			case coin:
				err = u.Coin()
				if err != nil {
					fmt.Printf("err: %s\n", err)
				}
				u.State.Event("toVal")
			case val:
				err = u.Val()
				if err != nil {
					fmt.Printf("err: %s\n", err)
				}
			default:
			}
		}
	}(fromClientChan)

	return fromClientChan, toClientChan
}

func NewUser(chatId int64) *User {
	u := &User{ChatID: chatId}

	u.State = fsm.NewFSM(
		menu,
		fsm.Events{
			{Name: "toAdd", Src: []string{add, coin, val, menu}, Dst: add},
			{Name: "toCoin", Src: []string{add}, Dst: coin},
			{Name: "toVal", Src: []string{coin}, Dst: val},
		},
		fsm.Callbacks{},
	)
	return u
}

func (u *User) Menu() string {
	return "menu"
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
