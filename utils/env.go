package utils

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var TgToken string
var PGUrl string
var CurrateKey string

func InitEnvVar(path ...string) (err error) {
	if len(path) > 0 {
		err = godotenv.Load(path[0])
	} else {
		err = godotenv.Load()
	}

	if err != nil {
		return
	}

	TgToken = os.Getenv("TOKEN")
	if TgToken == "" {
		return errors.New("empty Telegram Token")
	}

	PGUrl = os.Getenv("PG_URL")
	CurrateKey = os.Getenv("CURRATE_KEY")

	return
}
