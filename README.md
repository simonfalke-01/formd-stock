# FormD Stock Monitor

Efficient stock availability monitor for FormD products (Shopify-based stores). Polls the Shopify collection API and sends Telegram notifications when products come back in stock.

## Features

- **Efficient Polling**: ETag-based conditional requests minimize bandwidth
- **Smart Rate Limiting**: Exponential backoff with jitter prevents IP bans
- **State Tracking**: Detects new stock (unavailable → available transitions)
- **Telegram Notifications**: Real-time alerts with product details
- **Production Ready**: Graceful shutdown, structured logging, configurable

## Architecture

- **HTTP Client**: Optimized `net/http` with connection pooling and keep-alive
- **ETag Support**: `304 Not Modified` responses save ~95% bandwidth
- **State Manager**: Thread-safe variant availability tracking
- **Exponential Backoff**: Handles rate limits (429) and server errors (5xx)
- **Concurrent Safe**: Uses `sync.RWMutex` for state management

## Installation

```bash
# Clone the repository
git clone https://github.com/brandonli/formd-stock
cd formd-stock

# Install dependencies
go mod download

# Build
go build -o formd-stock
```

## Configuration

### Option 1: Environment Variables

```bash
# Copy example env file
cp .env.example .env

# Edit with your values
export SHOP_URL=https://formdt1.com
export TELEGRAM_TOKEN=your_bot_token
export TELEGRAM_CHAT_ID=your_chat_id
export POLL_INTERVAL=15s

# Run
./formd-stock
```

### Option 2: Config File

```bash
# Copy example config
cp config.json.example config.json

# Edit config.json with your values
nano config.json

# Run with config file
./formd-stock -config config.json
```

## Getting Telegram Credentials

1. **Create Bot**:
   - Message [@BotFather](https://t.me/botfather) on Telegram
   - Send `/newbot` and follow prompts
   - Save the bot token

2. **Get Chat ID**:
   - Message [@userinfobot](https://t.me/userinfobot)
   - It will reply with your chat ID

## Usage

```bash
# Run with environment variables
./formd-stock

# Run with config file
./formd-stock -config config.json

# Build and run
go run .

# Build for production
go build -ldflags="-s -w" -o formd-stock
```

## Configuration Options

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SHOP_URL` | Shop base URL | - | Yes |
| `COLLECTION_PATH` | Collection API path | `/collections/all/products.json?limit=250` | No |
| `POLL_INTERVAL` | Polling interval | `15s` | No |
| `TELEGRAM_TOKEN` | Telegram bot token | - | Yes* |
| `TELEGRAM_CHAT_ID` | Telegram chat ID | - | Yes* |
| `USER_AGENT` | HTTP User-Agent | `FormD-Stock-Monitor/1.0` | No |

\* Required for notifications (monitor will run without them but won't notify)

## Performance Optimizations

1. **ETag Caching**: Only downloads data when changed
2. **Connection Pooling**: Reuses HTTP connections
3. **Keep-Alive**: Maintains persistent connections
4. **HTTP/2**: Automatic upgrade when available
5. **Brotli/Gzip**: Automatic compression support

## Rate Limiting Protection

- Default interval: 15 seconds (safe for Shopify)
- Detects `429 Too Many Requests`
- Exponential backoff: 15s → 30s → 60s → 120s → 300s (max)
- Adds ±20% jitter to prevent thundering herd

## Output Example

```
2025/10/17 15:20:00 Loaded config from environment variables
2025/10/17 15:20:00 Authorized on Telegram as FormDStockBot
2025/10/17 15:20:00 Starting monitor for https://formdt1.com/collections/all/products.json?limit=250
2025/10/17 15:20:00 Poll interval: 15s
2025/10/17 15:20:01 Fetched 12 products
2025/10/17 15:20:16 No changes (304 Not Modified)
2025/10/17 15:20:31 No changes (304 Not Modified)
2025/10/17 15:20:46 Fetched 12 products
2025/10/17 15:20:46 Detected 1 stock changes
2025/10/17 15:20:46 NEW STOCK: T1 Customize - Version 2.1 / Anodized Black / Steel Coated ($195.00)
2025/10/17 15:20:46 Sent batch notification for 1 items
```

## Development

```bash
# Run tests
go test -v ./...

# Format code
go fmt ./...

# Lint
golangci-lint run

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o formd-stock-linux
GOOS=darwin GOARCH=arm64 go build -o formd-stock-macos
GOOS=windows GOARCH=amd64 go build -o formd-stock.exe
```

## Deployment

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w" -o formd-stock

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/formd-stock /formd-stock
CMD ["/formd-stock"]
```

### systemd Service

```ini
[Unit]
Description=FormD Stock Monitor
After=network.target

[Service]
Type=simple
User=nobody
EnvironmentFile=/etc/formd-stock/config.env
ExecStart=/usr/local/bin/formd-stock
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Monitoring Other Shopify Stores

This monitor works with any Shopify store:

```bash
export SHOP_URL=https://your-store.myshopify.com
export COLLECTION_PATH=/collections/all/products.json?limit=250
```

## Troubleshooting

**Rate Limited (429)**:
- Increase `POLL_INTERVAL` to 30s or higher
- Monitor will automatically back off

**No Notifications**:
- Verify bot token: `curl https://api.telegram.org/bot<TOKEN>/getMe`
- Check chat ID is correct
- Ensure bot can send messages to the chat

**ETag Not Working**:
- Some CDNs strip ETags - monitor will still work, just less efficient
- Check response headers with `curl -I <URL>`

## License

MIT

## Credits

Built for monitoring FormD T1 case restocks.
