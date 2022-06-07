package api

import (
	"testing"
	"wallet_tgbot/utils"

	"github.com/stretchr/testify/require"
)

func TestGetPrice(t *testing.T) {
	utils.InitEnvVar("../.env")
	for _, v := range []string{"btc", "eth"} {
		price, priceUSDRUB, err := GetPrice(v)
		t.Logf("price: %f priceUSDRUB: %f", price, priceUSDRUB)
		require.Nil(t, err)
	}
}
