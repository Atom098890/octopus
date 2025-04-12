# Tech News Telegram Bot

Telegram bot that delivers daily technology news summaries with ChatGPT-powered analysis.

## Features

- Daily technology news updates from NewsAPI
- AI-powered article summarization using ChatGPT
- Keyword extraction and Russian translation
- Beautifully formatted Telegram messages
- Containerized deployment with Docker

## Prerequisites

- Go 1.21 or later
- Docker (for containerized deployment)
- API Keys:
  - Telegram Bot Token
  - NewsAPI.org API Key
  - OpenAI API Key
  - Telegram Chat ID

## Local Development Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/goBot.git
cd goBot
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up environment variables:
```bash
export TELEGRAM_BOT_TOKEN="your_token"
export TELEGRAM_CHAT_ID="your_chat_id"
export NEWS_API_KEY="your_newsapi_key"
export OPENAI_API_KEY="your_openai_key"
```

4. Run the application:
```bash
go run cmd/bot/main.go
```

## Docker Deployment

1. Build the Docker image:
```bash
docker build -t tech-news-bot .
```

2. Run the container:
```bash
docker run -d \
  -e TELEGRAM_BOT_TOKEN="your_token" \
  -e TELEGRAM_CHAT_ID="your_chat_id" \
  -e NEWS_API_KEY="your_newsapi_key" \
  -e OPENAI_API_KEY="your_openai_key" \
  tech-news-bot
```

## GitHub Actions CI/CD

The project includes a GitHub Actions workflow that:
1. Runs tests
2. Builds the Docker image
3. Pushes to GitHub Container Registry
4. Deploys to production (when merged to main)

To use GitHub Actions:
1. Add your secrets in GitHub repository settings
2. Push to the main branch to trigger the workflow

## Project Structure

```
/
├── cmd/
│   └── bot/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/              # Configuration handling
│   ├── news/                # NewsAPI integration
│   ├── summarizer/          # ChatGPT integration
│   └── telegram/            # Telegram bot logic
├── Dockerfile               # Docker configuration
├── .github/workflows/       # CI/CD configuration
└── README.md               # This file
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request 