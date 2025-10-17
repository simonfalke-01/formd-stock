package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramNotifier sends notifications via Telegram
type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
	shopURL string
}

// NewTelegramNotifier creates a new Telegram notifier
func NewTelegramNotifier(token string, chatID int64, shopURL string) (*TelegramNotifier, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("creating telegram bot: %w", err)
	}

	log.Printf("Authorized on Telegram as %s", bot.Self.UserName)

	return &TelegramNotifier{
		bot:     bot,
		chatID:  chatID,
		shopURL: shopURL,
	}, nil
}

// NotifyStockChange sends a notification for a stock change
func (tn *TelegramNotifier) NotifyStockChange(change StockChange) error {
	if !change.IsNewStock() {
		return nil // Only notify on new stock
	}

	productURL := fmt.Sprintf("%s/products/%s",
		strings.TrimSuffix(tn.shopURL, "/"),
		change.ProductHandle,
	)

	// Format message with Markdown
	message := fmt.Sprintf(
		"🚨 *STOCK ALERT* 🚨\n\n"+
		"*Product:* %s\n"+
		"*Variant:* %s\n"+
		"*Price:* $%s\n"+
		"*SKU:* %s\n\n"+
		"[🛒 Buy Now](%s)",
		escapeMarkdown(change.ProductTitle),
		escapeMarkdown(change.VariantTitle),
		change.VariantPrice,
		change.VariantSKU,
		productURL,
	)

	msg := tgbotapi.NewMessage(tn.chatID, message)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = false

	_, err := tn.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sending telegram message: %w", err)
	}

	log.Printf("Sent notification for %s - %s", change.ProductTitle, change.VariantTitle)
	return nil
}

// NotifyMultiple sends a batch notification for multiple stock changes
func (tn *TelegramNotifier) NotifyMultiple(changes []StockChange) error {
	if len(changes) == 0 {
		return nil
	}

	// Filter to only new stock
	var newStock []StockChange
	for _, change := range changes {
		if change.IsNewStock() {
			newStock = append(newStock, change)
		}
	}

	if len(newStock) == 0 {
		return nil
	}

	// Build message
	var sb strings.Builder
	sb.WriteString("🚨 *STOCK ALERT* 🚨\n\n")
	sb.WriteString(fmt.Sprintf("*%d item(s) now in stock:*\n\n", len(newStock)))

	for i, change := range newStock {
		productURL := fmt.Sprintf("%s/products/%s",
			strings.TrimSuffix(tn.shopURL, "/"),
			change.ProductHandle,
		)

		sb.WriteString(fmt.Sprintf(
			"%d\\. *%s*\n"+
			"   Variant: %s\n"+
			"   Price: $%s\n"+
			"   [Buy Now](%s)\n\n",
			i+1,
			escapeMarkdown(change.ProductTitle),
			escapeMarkdown(change.VariantTitle),
			change.VariantPrice,
			productURL,
		))
	}

	msg := tgbotapi.NewMessage(tn.chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = false

	_, err := tn.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("sending telegram message: %w", err)
	}

	log.Printf("Sent batch notification for %d items", len(newStock))
	return nil
}

// SendMessage sends a plain text message
func (tn *TelegramNotifier) SendMessage(text string) error {
	msg := tgbotapi.NewMessage(tn.chatID, text)
	_, err := tn.bot.Send(msg)
	return err
}

// escapeMarkdown escapes special characters for Telegram MarkdownV1
func escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"`", "\\`",
	)
	return replacer.Replace(text)
}
