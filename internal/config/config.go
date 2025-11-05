package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Environment struct {
	ApiKey                   string `envconfig:"API_KEY"`
	ViewDnsApiUrl            string `envconfig:"VIEW_DNS_API_URL"`
	DBHost                   string `envconfig:"DB_HOST" default:"localhost"`
	DBPort                   int    `envconfig:"DB_PORT" default:"3306"`
	DBDatabase               string `envconfig:"DB_NAME"`
	DBUsername               string `envconfig:"DB_USER"`
	DBPassword               string `envconfig:"DB_PASSWORD"`
	RedisHost                string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort                int    `envconfig:"REDIS_PORT" default:"6379"`
	Address                  string `envconfig:"ADDRESS" default:"localhost"`
	Password                 string `envconfig:"PASSWORD"`
	JWTSecret                string `envconfig:"JWT_SECRET"`
	OpenaiApiKey             string `envconfig:"OPENAI_API_KEY"`
	SwaggerHost              string `envconfig:"SWAGGER_HOST" default:"localhost:8091"`
	NoticeWebAppChannelUrl   string `envconfig:"NOTICE_WEB_APP_CHANNEL_URL"`
	GoogleServiceAccountPath string `envconfig:"GOOGLE_SERVICE_ACCOUNT_PATH"`
	DatabasePassword1        string `envconfig:"DATABASE_PASSWORD_1"`
	DatabasePassword2        string `envconfig:"DATABASE_PASSWORD_2"`
	DatabaseHost1            string `envconfig:"DATABASE_HOST_1" default:"localhost"`
	DatabaseHost2            string `envconfig:"DATABASE_HOST_2" default:"localhost"`
	HashPhrase               string `envconfig:"HASH_PHRASE"`
	RodutSecretPhrase        string `envconfig:"RODUT_SECRET_PHRASE"`
	SheetID                  string `envconfig:"SHEET_ID"`
}

var Env Environment

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(".env loading failed")
	}
	err = envconfig.Process("", &Env)
	if err != nil {
		panic(err)
	}
}
