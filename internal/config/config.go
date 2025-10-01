package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Environment struct {
	ApiKey        string `envconfig:"API_KEY"`
	ViewDnsApiUrl string `envconfig:"VIEW_DNS_API_URL"`
	DBHost        string `envconfig:"DB_HOST" default:"localhost"`
	DBPort        int    `envconfig:"DB_PORT" default:"3306"`
	DBDatabase    string `envconfig:"DB_NAME"`
	DBUsername    string `envconfig:"DB_USER"`
	DBPassword    string `envconfig:"DB_PASSWORD"`
	Address       string `envconfig:"ADDRESS" default:"localhost"`
	Password      string `envconfig:"PASSWORD" required:"true"`
	JWTSecret     string `envconfig:"JWT_SECRET" required:"true"`
}

var Env Environment

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	err = envconfig.Process("", &Env)
	if err != nil {
		panic(err)
	}
}
