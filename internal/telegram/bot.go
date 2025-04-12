package telegram

import (
	"fmt"
	"log"
	"strings"

	"github.com/andrei/goBot/internal/news"
	"github.com/andrei/goBot/internal/summarizer"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	users  *Users
	logger *log.Logger
}

func NewBot(token string, logger *log.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("error creating telegram bot: %w", err)
	}

	return &Bot{
		api:    api,
		users:  NewUsers(),
		logger: logger,
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	// Пропускаем обновления, которые могли накопиться во время простоя бота
	// Но проверяем, не была ли среди них команда /start
	for update := range updates {
		if update.Message != nil && update.Message.Command() == "start" {
			b.handleStartCommand(update.Message)
		}
		
		// Проверяем, пуст ли канал
		if len(updates) == 0 {
			break
		}
	}

	b.logger.Println("Bot started and ready to receive messages")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Обработка команд
		switch update.Message.Command() {
		case "start":
			b.handleStartCommand(update.Message)
		case "news":
			// Отправляем сообщение о том, что обрабатываем запрос
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Получаю последние технологические новости...")
			b.api.Send(msg)
			
			// Здесь можно вызвать processNews, но так как функция находится в main.go,
			// просто отправляем уведомление
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Функция обработки новостей по запросу будет добавлена в следующем обновлении. Пока новости приходят по расписанию.")
			b.api.Send(msg)
		case "help":
			b.handleHelpCommand(update.Message)
		}
	}
}

// handleStartCommand обрабатывает команду /start
func (b *Bot) handleStartCommand(message *tgbotapi.Message) {
	userID := message.Chat.ID
	userName := message.From.UserName
	
	// Добавляем пользователя в список подписчиков
	b.users.Add(userID)
	
	// Формируем приветственное сообщение
	greeting := fmt.Sprintf("Привет, %s! 👋\n\n"+
		"Я бот технологических новостей. Я буду присылать тебе интересные новости из мира технологий.\n\n"+
		"<b>Доступные команды:</b>\n"+
		"/start - Запустить бота\n"+
		"/news - Получить последние новости\n"+
		"/help - Показать помощь\n\n"+
		"Жди первую новость или используй команду /news, чтобы получить её сейчас!",
		userName)
	
	msg := tgbotapi.NewMessage(userID, greeting)
	msg.ParseMode = "HTML"
	b.api.Send(msg)
	
	b.logger.Printf("New user subscribed: %s (ID: %d)", userName, userID)
}

// handleHelpCommand обрабатывает команду /help
func (b *Bot) handleHelpCommand(message *tgbotapi.Message) {
	helpText := "<b>Помощь по использованию бота:</b>\n\n" +
		"Этот бот отправляет технологические новости по расписанию.\n\n" +
		"<b>Доступные команды:</b>\n" +
		"/start - Запустить бота и подписаться на новости\n" +
		"/news - Получить последние новости сейчас\n" +
		"/help - Показать эту помощь\n\n" +
		"Если у вас возникли проблемы, пожалуйста, свяжитесь с разработчиком."
	
	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
	msg.ParseMode = "HTML"
	b.api.Send(msg)
}

func (b *Bot) SendArticleSummary(article *news.Article, summary *summarizer.Summary) error {
	message := b.formatMessage(article, summary)
	
	for _, userID := range b.users.GetAll() {
		msg := tgbotapi.NewMessage(userID, message)
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = false

		_, err := b.api.Send(msg)
		if err != nil {
			b.logger.Printf("Error sending message to user %d: %v", userID, err)
			continue
		}
	}

	return nil
}

func (b *Bot) formatMessage(article *news.Article, summary *summarizer.Summary) string {
	var sb strings.Builder

	// Заголовок статьи
	sb.WriteString(fmt.Sprintf("<b>📰 %s</b>\n", article.Title))
	
	// Источник и автор
	if article.Source.Name != "" {
		sb.WriteString(fmt.Sprintf("📢 <i>%s</i>", article.Source.Name))
		if article.Author != "" {
			sb.WriteString(fmt.Sprintf(" | ✍️ %s", article.Author))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Текст статьи (обрезаем до разумного размера)
	content := article.Content
	if len(content) > 800 {
		lastDot := strings.LastIndex(content[:800], ".")
		if lastDot > 0 {
			content = content[:lastDot+1]
		} else {
			content = content[:800] + "..."
		}
	}
	sb.WriteString(content)
	sb.WriteString("\n\n")

	// Ключевые термины и их перевод
	sb.WriteString("<b>🔑 Key Terms:</b>\n")
	for _, keyword := range summary.Keywords {
		translation, ok := summary.Translation[keyword]
		if ok {
			sb.WriteString(fmt.Sprintf("• %s — %s\n", keyword, translation))
		}
	}
	sb.WriteString("\n")

	// Ссылка на оригинал
	sb.WriteString(fmt.Sprintf("🔗 <a href=\"%s\">Read full article</a>\n", article.URL))

	// Дата публикации
	sb.WriteString(fmt.Sprintf("\n📅 Published: %s", article.PublishedAt.Format("02.01.2006 15:04")))

	return sb.String()
}

// Users возвращает объект пользователей
func (b *Bot) Users() *Users {
	return b.users
}