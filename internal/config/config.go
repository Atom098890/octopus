package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken string
	NewsAPIKey       string
	OpenAIAPIKey     string
	NewsCategory     string
	NewsLanguage     string
	ScheduleTime     string
}

func LoadConfig() (*Config, error) {
	// Настройка Viper для чтения из файла
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Чтение файла
	if err := viper.ReadInConfig(); err != nil {
		// Если файл не найден, пробуем читать из переменных окружения
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Установка значений по умолчанию
	viper.SetDefault("NEWS_CATEGORY", "technology")
	viper.SetDefault("NEWS_LANGUAGE", "en")
	viper.SetDefault("SCHEDULE_TIME", "0 9 * * *") // По умолчанию в 9:00 каждый день

	// Чтение из переменных окружения
	viper.AutomaticEnv()

	// Проверка обязательных переменных
	requiredEnvs := []string{
		"TELEGRAM_BOT_TOKEN",
		"NEWS_API_KEY",
		"OPENAI_API_KEY",
	}

	for _, env := range requiredEnvs {
		if !viper.IsSet(env) {
			return nil, fmt.Errorf("required environment variable %s is not set", env)
		}
	}

	return &Config{
		TelegramBotToken: viper.GetString("TELEGRAM_BOT_TOKEN"),
		NewsAPIKey:       viper.GetString("NEWS_API_KEY"),
		OpenAIAPIKey:     viper.GetString("OPENAI_API_KEY"),
		NewsCategory:     viper.GetString("NEWS_CATEGORY"),
		NewsLanguage:     viper.GetString("NEWS_LANGUAGE"),
		ScheduleTime:     viper.GetString("SCHEDULE_TIME"),
	}, nil
} 