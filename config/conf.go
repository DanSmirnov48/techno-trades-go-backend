package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	EmailOtpExpireMins        int64  `mapstructure:"EMAIL_OTP_EXPIRE_MINS"`
	AccessTokenExpireMinutes  int    `mapstructure:"ACCESS_TOKEN_EXPIRE_MINUTES"`
	RefreshTokenExpireMinutes int    `mapstructure:"REFRESH_TOKEN_EXPIRE_MINUTES"`
	Port                      string `mapstructure:"PORT"`
	SecretKey                 string `mapstructure:"SECRET_KEY"`
	PostgresUser              string `mapstructure:"POSTGRES_USER"`
	PostgresPassword          string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresServer            string `mapstructure:"POSTGRES_SERVER"`
	PostgresPort              string `mapstructure:"POSTGRES_PORT"`
	PostgresDB                string `mapstructure:"POSTGRES_DB"`
	TestPostgresDB            string `mapstructure:"TEST_POSTGRES_DB"`
	MailSenderEmail           string `mapstructure:"MAIL_SENDER_EMAIL"`
	MailSenderPassword        string `mapstructure:"MAIL_SENDER_PASSWORD"`
	MailSenderHost            string `mapstructure:"MAIL_SENDER_HOST"`
	MailSenderPort            int    `mapstructure:"MAIL_SENDER_PORT"`
	CORSAllowedOrigins        string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	FrontendURL               string `mapstructure:"CLIENT_URL"`
	StripeTestKey             string `mapstructure:"STRIPE_TEST_KEY"`
	StripeSecretKey           string `mapstructure:"STRIPE_SECRET_KEY"`
	GoogleClientId            string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret        string `mapstructure:"GOOGLE_CLIENT_SECRET"`
}

func GetConfig(testOpts ...bool) (config Config) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	var err error
	if err = viper.ReadInConfig(); err != nil {
		panic(err)
	}
	viper.Unmarshal(&config)
	return
}
