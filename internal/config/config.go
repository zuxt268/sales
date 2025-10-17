package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Environment struct {
	ApiKey                 string `envconfig:"API_KEY"`
	ViewDnsApiUrl          string `envconfig:"VIEW_DNS_API_URL"`
	DBHost                 string `envconfig:"DB_HOST" default:"localhost"`
	DBPort                 int    `envconfig:"DB_PORT" default:"3306"`
	DBDatabase             string `envconfig:"DB_NAME"`
	DBUsername             string `envconfig:"DB_USER"`
	DBPassword             string `envconfig:"DB_PASSWORD"`
	Address                string `envconfig:"ADDRESS" default:"localhost"`
	Password               string `envconfig:"PASSWORD"`
	JWTSecret              string `envconfig:"JWT_SECRET"`
	OpenaiApiKey           string `envconfig:"OPENAI_API_KEY"`
	SwaggerHost            string `envconfig:"SWAGGER_HOST" default:"localhost:8091"`
	NoticeWebAppChannelUrl string `envconfig:"NOTICE_WEB_APP_CHANNEL_URL"`
}

var Env Environment

func init() {
	err := envconfig.Process("", &Env)
	if err != nil {
		panic(err)
	}
}
