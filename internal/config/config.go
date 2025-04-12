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
	// Настройка Viper для чтения из переменных окружения
	viper.AutomaticEnv()

	// Также попробуем прочитать из файла .env, если он существует
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	
	// Чтение файла
	if err := viper.ReadInConfig(); err != nil {
		// Если файл не найден, продолжаем работу с переменными окружения
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Если ошибка не связана с отсутствием файла, логируем её,
			// но продолжаем работу с переменными окружения
			fmt.Printf("Warning: Error reading .env file: %v\n", err)
		}
	}

	// Установка значений по умолчанию
	viper.SetDefault("NEWS_CATEGORY", "technology")
	viper.SetDefault("NEWS_LANGUAGE", "en")
	viper.SetDefault("SCHEDULE_TIME", "0 9 * * *") // По умолчанию в 9:00 каждый день

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