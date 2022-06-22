package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"wallet_tgbot/api"
	"wallet_tgbot/utils"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

type wallet struct {
	ChatId   int64   `db:"chat_id"`
	Currency string  `db:"currency"`
	Value    float64 `db:"value"`
	conn     *sqlx.DB
}

var (
	ErrValLessZero = errors.New("вычитаемое значение больше суммы в кошельке")
	ErrNoRowsToDel = errors.New("нет валюты для удаления")
)

func NewWallet(chat_id int64) wallet {
	return wallet{ChatId: chat_id}
}

func (w *wallet) connect() error {
	conn, err := sqlx.Connect("pgx", utils.PGUrl)
	if err != nil {
		log.Println(err)
		return err
	}

	w.conn = conn

	return nil
}

func (w *wallet) Add(coin string, val float64) (float64, error) {
	err := w.connect()
	if err != nil {
		return 0, err
	}
	defer w.conn.Close()

	query := "update wallet set value=:value where chat_id=:chat_id and currency=:currency"

	coin = strings.ToLower(coin)
	temp_w := wallet{ChatId: w.ChatId, Currency: coin, Value: 0}

	err = w.conn.Get(
		&temp_w, "select value from wallet where chat_id = $1 and currency = $2",
		temp_w.ChatId, temp_w.Currency,
	)
	if err == sql.ErrNoRows {
		query = "insert into wallet (chat_id, currency, value) values (:chat_id, :currency, :value)"
	} else if err != nil {
		return 0, err
	}

	temp_w.Value += val
	_, err = w.conn.NamedExec(query, temp_w)
	if err != nil {
		return 0, err
	}

	return temp_w.Value, nil
}

func (w *wallet) Sub(coin string, val float64) (float64, error) {
	err := w.connect()
	if err != nil {
		return 0, err
	}
	defer w.conn.Close()

	temp_w := wallet{ChatId: w.ChatId, Currency: coin, Value: 0}

	err = w.conn.Get(
		&temp_w, "select value from wallet where chat_id = $1 and currency = $2",
		temp_w.ChatId, temp_w.Currency,
	)
	if err == sql.ErrNoRows {
		return 0, ErrValLessZero
	} else if err != nil {
		return 0, err
	}

	if temp_w.Value-val < 0 {
		return 0, ErrValLessZero
	}
	temp_w.Value -= val
	query := "update wallet set value=:value where chat_id=:chat_id and currency=:currency"
	_, err = w.conn.NamedExec(query, temp_w)
	if err != nil {
		return 0, err
	}

	return temp_w.Value, nil
}

func (w *wallet) Delete(coin string) error {
	err := w.connect()
	if err != nil {
		return err
	}
	defer w.conn.Close()

	query := "delete from wallet where chat_id=$1 and currency=$2"
	res, err := w.conn.Exec(query, w.ChatId, coin)
	if err != nil {
		return err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return ErrNoRowsToDel
	}

	return nil
}

func (w *wallet) Show() (string, error) {
	err := w.connect()
	if err != nil {
		return "", err
	}
	defer w.conn.Close()

	var temp_w []wallet
	err = w.conn.Select(&temp_w, "select currency, value from wallet where chat_id = $1", w.ChatId)
	if err != nil {
		return "", err
	}

	if len(temp_w) == 0 {
		return "Нет валют", nil
	}
	balance := "Баланс:\n"

	for _, v := range temp_w {
		currencyUpper := strings.ToUpper(v.Currency)
		price, priceUSDRUB, err := api.GetPrice(currencyUpper)
		s := ""
		if errors.Is(err, api.ErrGetRateCoin) {
			s = fmt.Sprintf("`%-4s`: %f \\[ - USD - RUB ]\n", currencyUpper, v.Value)
		} else if errors.Is(err, api.ErrGetRateRUB) {
			s = fmt.Sprintf("`%-4s`: %f \\[ %.2f USD - RUB ]\n", currencyUpper, v.Value, price*v.Value)
		} else if err != nil {
			s = fmt.Sprintf("`%-4s`: %f \\[ - USD - RUB ]\n", currencyUpper, v.Value)
		} else {
			s = fmt.Sprintf(
				"`%-4s`: %f \\[%.2f USD %.2f RUB ]\n",
				currencyUpper, v.Value, price*v.Value, price*v.Value*priceUSDRUB)
		}

		if err != nil {
			utils.Loggers.Errorw(
				"ошибка получения курса",
				"chat_id", w.ChatId,
				"err", err,
			)
		}

		balance += s
	}

	return balance, nil
}

func (w *wallet) GetCurrency() ([]string, error) {
	err := w.connect()
	if err != nil {
		return []string{}, err
	}
	defer w.conn.Close()

	var temp_w []wallet
	err = w.conn.Select(&temp_w, "select upper(currency) as currency from wallet where chat_id = $1", w.ChatId)
	if err != nil {
		return []string{}, err
	}

	res := make([]string, len(temp_w))
	for i, v := range temp_w {
		res[i] = v.Currency
	}

	return res, nil
}
