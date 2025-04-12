package news

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Article struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
	Source      struct {
		Name string `json:"name"`
	} `json:"source"`
	Author string `json:"author"`
}

type NewsAPIResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) FetchLatestTechNews(language string) (*Article, error) {
	// Получаем несколько статей для выбора лучшей
	articles, err := c.fetchArticles(language, 10)
	if err != nil {
		return nil, err
	}

	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles found")
	}

	// Выбираем лучшую статью на основе длины контента и наличия важных полей
	bestArticle := c.selectBestArticle(articles)

	// Дополняем контент описанием, если он короткий
	if len(bestArticle.Content) < len(bestArticle.Description) {
		bestArticle.Content = bestArticle.Description + "\n\n" + bestArticle.Content
	}

	// Очищаем контент от технических артефактов
	bestArticle.Content = c.cleanContent(bestArticle.Content)

	return bestArticle, nil
}

func (c *Client) fetchArticles(language string, pageSize int) ([]Article, error) {
	// Используем everything endpoint вместо top-headlines для большего охвата
	baseURL := "https://newsapi.org/v2/everything"
	
	params := url.Values{}
	params.Add("language", language)
	params.Add("pageSize", fmt.Sprintf("%d", pageSize))
	params.Add("sortBy", "publishedAt")

	// Добавляем основные технологические домены
	domains := []string{
		"techcrunch.com",
		"theverge.com",
		"wired.com",
		"arstechnica.com",
		"engadget.com",
		"zdnet.com",
		"venturebeat.com",
		"thenextweb.com",
	}
	params.Add("domains", strings.Join(domains, ","))
	
	// Используем только базовый поиск для начала
	params.Add("q", "technology")

	requestURL := baseURL + "?" + params.Encode()
	fmt.Printf("Making request to: %s\n", requestURL)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("news API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	fmt.Printf("Found %d articles\n", len(apiResp.Articles))

	if len(apiResp.Articles) == 0 {
		// Если статьи не найдены, пробуем более широкий поиск
		params.Del("domains") // Убираем ограничение по доменам
		
		requestURL = baseURL + "?" + params.Encode()
		fmt.Printf("Making fallback request to: %s\n", requestURL)

		req, err = http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating fallback request: %w", err)
		}

		req.Header.Set("X-Api-Key", c.apiKey)

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error making fallback request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("news API returned non-200 status code in fallback: %d, body: %s", resp.StatusCode, string(body))
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			return nil, fmt.Errorf("error decoding fallback response: %w", err)
		}
		
		fmt.Printf("Found %d articles in fallback request\n", len(apiResp.Articles))
	}

	return apiResp.Articles, nil
}

func (c *Client) selectBestArticle(articles []Article) *Article {
	var bestArticle *Article
	var maxScore int

	for i, article := range articles {
		score := 0
		
		// Оценка длины контента
		score += len(article.Content) / 10
		score += len(article.Description) / 10
		
		// Бонус за наличие автора
		if article.Author != "" {
			score += 50
		}

		// Бонус за наличие источника
		if article.Source.Name != "" {
			score += 30
		}

		// Бонус за свежесть новости
		hoursAgo := time.Since(article.PublishedAt).Hours()
		if hoursAgo < 24 {
			score += 100
		} else if hoursAgo < 48 {
			score += 50
		}

		// Бонус за технологические ключевые слова в заголовке и описании
		techKeywords := []string{"technology", "tech", "software", "AI", "artificial intelligence", 
			"cybersecurity", "digital", "innovation", "startup", "algorithm", "cloud", "data", 
			"security", "privacy", "blockchain", "machine learning"}
		
		combinedText := strings.ToLower(article.Title + " " + article.Description)
		for _, keyword := range techKeywords {
			if strings.Contains(combinedText, strings.ToLower(keyword)) {
				score += 20
			}
		}

		// Выбираем статью с наивысшим счётом
		if bestArticle == nil || score > maxScore {
			bestArticle = &articles[i]
			maxScore = score
		}
	}

	return bestArticle
}

func (c *Client) cleanContent(content string) string {
	// Удаляем технические артефакты типа [+123 chars]
	content = strings.ReplaceAll(content, "chars]", "")
	content = strings.ReplaceAll(content, "[+", "")
	
	// Удаляем множественные пробелы и переносы строк
	content = strings.Join(strings.Fields(content), " ")
	
	// Добавляем переносы строк после точек для лучшей читаемости
	content = strings.ReplaceAll(content, ". ", ".\n\n")
	
	return content
} 