package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/andrei/goBot/internal/config"
	"github.com/andrei/goBot/internal/news"
	"github.com/andrei/goBot/internal/summarizer"
	"github.com/andrei/goBot/internal/telegram"
	"github.com/robfig/cron/v3"
)

func main() {
	// Инициализация логгера
	logger := log.New(os.Stdout, "TechNewsBot: ", log.LstdFlags)

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация компонентов
	newsClient := news.NewClient(cfg.NewsAPIKey)
	summarizer := summarizer.NewSummarizer(cfg.OpenAIAPIKey)
	bot, err := telegram.NewBot(cfg.TelegramBotToken, logger)
	if err != nil {
		logger.Fatalf("Failed to create Telegram bot: %v", err)
	}

	// Создание контекста с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WaitGroup для горутин
	var wg sync.WaitGroup

	// Запуск обработки сообщений бота
	wg.Add(1)
	go func() {
		defer wg.Done()
		bot.Start()
	}()

	// Инициализация планировщика
	c := cron.New()
	
	// Добавление задачи в планировщик
	_, err = c.AddFunc(cfg.ScheduleTime, func() {
		if err := processNews(ctx, newsClient, summarizer, bot, cfg, logger); err != nil {
			logger.Printf("Error processing news: %v", err)
		}
	})
	if err != nil {
		logger.Fatalf("Failed to schedule task: %v", err)
	}

	// Запуск планировщика
	c.Start()
	logger.Printf("Bot started. Scheduled to run at %s", cfg.ScheduleTime)

	// Обработка сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down...")
	c.Stop()
	cancel()
	wg.Wait()
}

func processNews(ctx context.Context, newsClient *news.Client, summarizer *summarizer.Summarizer, bot *telegram.Bot, cfg *config.Config, logger *log.Logger) error {
	// Получение последней новости
	article, err := newsClient.FetchLatestTechNews(cfg.NewsLanguage)
	if err != nil {
		return fmt.Errorf("failed to fetch news: %w", err)
	}

	// Обработка статьи через ChatGPT
	summary, err := summarizer.ProcessArticle(ctx, article)
	if err != nil {
		return fmt.Errorf("failed to process article: %w", err)
	}

	// Отправка сообщения в Telegram
	if err := bot.SendArticleSummary(article, summary); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	logger.Printf("Successfully processed and sent article: %s", article.Title)
	return nil
} 