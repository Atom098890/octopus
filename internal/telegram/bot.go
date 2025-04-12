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
	// Очищаем предыдущие обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)
	
	// Пропускаем все старые сообщения
	for len(updates) > 0 {
		<-updates
	}

	b.logger.Println("Bot started and ready to receive messages")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Command() == "start" {
			b.users.Add(update.Message.Chat.ID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я буду присылать тебе дайджест технологических новостей. Жди первую новость!")
			b.api.Send(msg)
		}
	}
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