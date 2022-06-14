package model

import (
	"database/sql"
	"testing"
	"wallet_tgbot/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func delete_data() (err error) {
	w := wallet{}
	err = w.connect()
	if err != nil {
		return
	}
	defer w.conn.Close()

	w.conn.Exec("delete from wallet")
	return
}

func insert_data(ws []wallet) (err error) {
	w := wallet{}
	err = w.connect()
	if err != nil {
		return
	}
	defer w.conn.Close()

	for _, v := range ws {
		_, err = w.conn.NamedExec(
			"insert into wallet (chat_id, currency, value) values (:chat_id, :currency, :value)",
			v,
		)
		if err != nil {
			return
		}
	}
	return
}

func TestConn(t *testing.T) {
	utils.InitEnvVar("../.env")

	w := wallet{}
	err := w.connect()
	require.Nil(t, err)
	w.conn.Close()
}

func TestAdd(t *testing.T) {
	utils.InitEnvVar("../.env")
	err := delete_data()
	require.Nil(t, err)

	testCase := []struct {
		ChatId   int64
		Currency string
		Values   []float64
	}{
		{1, "btc", []float64{0.1, 0.1, 0.2}},
		{1, "eth", []float64{0.1, 1.1, 0.2}},
		{1, "eth", []float64{0.4, 1.1, 0.2}},
		{2, "eth", []float64{0, 0, 0}},
	}

	for _, tc := range testCase {
		chat_id := tc.ChatId
		currency := tc.Currency
		vals := tc.Values

		for _, val := range vals {

			old_val, new_val, err := add_t(chat_id, currency, val)

			require.Nil(t, err)
			assert.Equal(t, old_val+val, new_val, "Val after Add must be equal")
		}
	}
}

func TestSub(t *testing.T) {
	utils.InitEnvVar("../.env")
	err := delete_data()
	require.Nil(t, err)

	err = insert_data([]wallet{
		{ChatId: 1, Currency: "btc", Value: 2.6},
		{ChatId: 1, Currency: "eth", Value: 34.6},
		{ChatId: 2, Currency: "btc", Value: 0.6},
	})
	require.Nil(t, err)

	testCase := []struct {
		ChatId   int64
		Currency string
		Values   []float64
		Err      []error
	}{
		{1, "btc", []float64{0.1, 0.1, 0.2}, []error{nil, nil, nil}},
		{1, "eth", []float64{0.1, 1.1, 60.2}, []error{nil, nil, ErrValLessZero}},
		{2, "btc", []float64{0, 0, 0}, []error{nil, nil, nil}},
		{2, "eth", []float64{0, 16, 0.3}, []error{ErrValLessZero, ErrValLessZero, ErrValLessZero}},
	}

	for _, tc := range testCase {
		chat_id := tc.ChatId
		currency := tc.Currency
		vals := tc.Values

		for i, val := range vals {

			old_val, new_val, err := sub_t(chat_id, currency, val)

			if err != nil {
				require.ErrorIs(t, err, tc.Err[i])
			} else {
				assert.Equal(t, old_val-val, new_val, "Val after sub must be equal")
			}
		}
	}
}

func TestDelete(t *testing.T) {
	utils.InitEnvVar("../.env")
	err := delete_data()
	require.Nil(t, err)

	err = insert_data([]wallet{
		{ChatId: 1, Currency: "btc", Value: 2.6},
		{ChatId: 1, Currency: "eth", Value: 34.6},
		{ChatId: 2, Currency: "btc", Value: 0.6},
		{ChatId: 3, Currency: "btc", Value: 4.6},
	})
	require.Nil(t, err)

	testCase := []struct {
		ChatId   int64
		Currency string
		Err      error
	}{
		{1, "btc", nil},
		{1, "eth", nil},
		{2, "btc", nil},
		{2, "eth", ErrNoRowsToDel},
	}

	for _, tc := range testCase {
		chat_id := tc.ChatId
		currency := tc.Currency

		err := del_t(chat_id, currency)

		require.ErrorIs(t, err, tc.Err)
	}
}

func TestShow(t *testing.T) {
	utils.InitEnvVar("../.env")
	err := delete_data()
	require.Nil(t, err)

	err = insert_data([]wallet{
		{ChatId: 1, Currency: "btc", Value: 2.6},
		{ChatId: 1, Currency: "eth", Value: 34.0007},
		{ChatId: 2, Currency: "btc", Value: 0.6},
		{ChatId: 3, Currency: "btc", Value: 4.6},
	})
	require.Nil(t, err)

	w := NewWallet(1)
	balance, err := w.Show()
	require.Nil(t, err)

	t.Log(balance)
}

func TestGetCurrency(t *testing.T) {
	utils.InitEnvVar("../.env")

	w := NewWallet(158911762)
	params, err := w.GetCurrency()
	require.Nil(t, err)

	t.Log(params)
}

func add_t(chat_id int64, currency string, val float64) (old_val, new_val float64, err error) {
	w := NewWallet(chat_id)

	err = w.connect()
	if err != nil {
		return
	}

	err = w.conn.QueryRow(
		"select value from wallet where chat_id=$1 and currency=$2", chat_id, currency,
	).Scan(&old_val)
	if err == sql.ErrNoRows {
		old_val = 0
	} else if err != nil {
		return
	}
	w.conn.Close()

	_, err = w.Add(currency, val)
	if err != nil {
		return
	}

	err = w.connect()
	if err != nil {
		return
	}

	err = w.conn.QueryRow(
		"select value from wallet where chat_id=$1 and currency=$2", chat_id, currency,
	).Scan(&new_val)
	if err != nil {
		return
	}
	w.conn.Close()

	return
}

func sub_t(chat_id int64, currency string, val float64) (old_val, new_val float64, err error) {
	w := NewWallet(chat_id)

	err = w.connect()
	if err != nil {
		return
	}

	err = w.conn.QueryRow(
		"select value from wallet where chat_id=$1 and currency=$2", chat_id, currency,
	).Scan(&old_val)
	if err == sql.ErrNoRows {
		err = ErrValLessZero
		return
	} else if err != nil {
		return
	}
	w.conn.Close()

	_, err = w.Sub(currency, val)
	if err != nil {
		return
	}

	err = w.connect()
	if err != nil {
		return
	}

	err = w.conn.QueryRow(
		"select value from wallet where chat_id=$1 and currency=$2", chat_id, currency,
	).Scan(&new_val)
	if err != nil {
		return
	}
	w.conn.Close()

	return
}

func del_t(chat_id int64, currency string) (err error) {
	w := NewWallet(chat_id)

	err = w.connect()
	if err != nil {
		return
	}
	defer w.conn.Close()

	err = w.Delete(currency)
	return
}
