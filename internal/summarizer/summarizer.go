package summarizer

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrei/goBot/internal/news"
	"github.com/sashabaranov/go-openai"
)

type Summary struct {
	Summary     string
	Keywords    []string
	Translation map[string]string
}

type Summarizer struct {
	client *openai.Client
}

func NewSummarizer(apiKey string) *Summarizer {
	return &Summarizer{
		client: openai.NewClient(apiKey),
	}
}

func (s *Summarizer) ProcessArticle(ctx context.Context, article *news.Article) (*Summary, error) {
	prompt := fmt.Sprintf(`Analyze this technology article and extract:

1. 5 key technical terms/concepts that are actually used in the article.
Rules for terms:
- Must be actual technology terminology (like "machine learning", "cloud computing", "neural network")
- Focus on technical concepts, tools, and methodologies
- Exclude company names, product names, and general words
- Terms should be 1-3 words long
- Each term must appear in the article text
- Prefer more specific technical terms over general ones

2. Provide accurate Russian translations for these technical terms

Article Title: %s
Article Content: %s

Format your response EXACTLY as follows (only these two lines):
Keywords: term1, term2, term3, term4, term5
Translations: term1: перевод1, term2: перевод2, term3: перевод3, term4: перевод4, term5: перевод5`, 
		article.Title, 
		article.Content)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error getting completion: %w", err)
	}

	return parseResponse(resp.Choices[0].Message.Content)
}

func parseResponse(response string) (*Summary, error) {
	parts := strings.Split(response, "\n")
	summary := &Summary{
		Keywords:    make([]string, 0),
		Translation: make(map[string]string),
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "Keywords:") {
			keywords := strings.TrimPrefix(part, "Keywords:")
			keywords = strings.TrimSpace(keywords)
			for _, keyword := range strings.Split(keywords, ",") {
				keyword = strings.TrimSpace(keyword)
				if keyword != "" {
					summary.Keywords = append(summary.Keywords, keyword)
				}
			}
		} else if strings.HasPrefix(part, "Translations:") {
			translations := strings.TrimPrefix(part, "Translations:")
			translations = strings.TrimSpace(translations)
			pairs := strings.Split(translations, ",")
			for _, pair := range pairs {
				pair = strings.TrimSpace(pair)
				kv := strings.Split(pair, ":")
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					if key != "" && value != "" {
						summary.Translation[key] = value
					}
				}
			}
		}
	}

	// Проверка на пустые ключевые слова
	if len(summary.Keywords) == 0 {
		return nil, fmt.Errorf("no keywords found in response: %s", response)
	}

	return summary, nil
} 