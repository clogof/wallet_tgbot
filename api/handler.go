package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"wallet_tgbot/utils"

	"github.com/tidwall/gjson"
)

var (
	ErrGetRateRUB  = errors.New("не удалось получить рублевый курс")
	ErrGetRateCoin = errors.New("не удалось получить крипто курс")
)

func GetPrice(coin string) (price float64, priceUSDRUB float64, err error) {
	base_url_wout_key := fmt.Sprintf(
		"https://currate.ru/api/?get=rates&pairs=USDRUB,%sUSD&key=", strings.ToUpper(coin),
	)
	base_url := base_url_wout_key + utils.CurrateKey

	c := http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", base_url, nil)
	if err != nil {
		return
	}

	resp, err := c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	body := string(bodyByte)

	g := gjson.Get(body, "status")
	if !g.Exists() || g.Int() != 200 {
		err = errors.New("Ошибка currate api:\n" + gjson.Get(body, "message").String())
		utils.Loggers.Errorw(
			"ошибка currate api",
			"url", base_url_wout_key+"####",
			"response", body,
			"err", err,
		)
		return
	}

	g = gjson.Get(body, fmt.Sprintf("data.%sUSD", strings.ToUpper(coin)))
	if !g.Exists() {
		err = ErrGetRateCoin
		utils.Loggers.Errorw(
			"ошибка currate api",
			"url", base_url_wout_key+"####",
			"response", body,
			"err", err,
		)
		return
	}
	price = g.Float()

	g = gjson.Get(body, "data.USDRUB")
	if !g.Exists() {
		err = ErrGetRateRUB
		utils.Loggers.Errorw(
			"ошибка currate api",
			"url", base_url_wout_key+"####",
			"response", body,
			"err", err,
		)
		return
	}
	priceUSDRUB = g.Float()

	return
}
