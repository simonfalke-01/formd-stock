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
		"*STOCK ALERT*\n\n"+
		"*Product:* [%s](%s)\n"+
		"*Variant:* `%s`\n"+
		"*Price:* *$%s*\n"+
		"*SKU:* `%s`\n"+
		"*Product ID:* `%d`\n"+
		"*Variant ID:* `%d`\n\n"+
		"*Transition:* OUT OF STOCK → IN STOCK",
		escapeMarkdown(change.ProductTitle),
		productURL,
		escapeMarkdown(change.VariantTitle),
		change.VariantPrice,
		change.VariantSKU,
		change.ProductID,
		change.VariantID,
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
	sb.WriteString("*STOCK ALERT*\n\n")
	sb.WriteString(fmt.Sprintf("*Detection:* %d item(s) transitioned to IN STOCK\n\n", len(newStock)))

	for i, change := range newStock {
		productURL := fmt.Sprintf("%s/products/%s",
			strings.TrimSuffix(tn.shopURL, "/"),
			change.ProductHandle,
		)

		sb.WriteString(fmt.Sprintf(
			"*%d\\.* [%s](%s)\n"+
			"     `%s`\n"+
			"     *$%s* • SKU: `%s`\n"+
			"     Product ID: `%d` • Variant ID: `%d`\n\n",
			i+1,
			escapeMarkdown(change.ProductTitle),
			productURL,
			escapeMarkdown(change.VariantTitle),
			change.VariantPrice,
			change.VariantSKU,
			change.ProductID,
			change.VariantID,
		))
	}

	sb.WriteString("State change detected via polling endpoint")

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

// SendStatusReport sends an initial status report
func (tn *TelegramNotifier) SendStatusReport(products []Product, totalVariants int, pollInterval string) error {
	var inStock, outOfStock int
	var inStockItems []string

	for _, product := range products {
		for _, variant := range product.Variants {
			if variant.Available {
				inStock++
				productURL := fmt.Sprintf("%s/products/%s",
					strings.TrimSuffix(tn.shopURL, "/"),
					product.Handle,
				)
				inStockItems = append(inStockItems, fmt.Sprintf(
					"  • [%s](%s)\n    `%s` — *$%s*\n    ID: `%d` • SKU: `%s`",
					escapeMarkdown(product.Title),
					productURL,
					escapeMarkdown(variant.Title),
					variant.Price,
					variant.ID,
					variant.SKU,
				))
			} else {
				outOfStock++
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("*STOCK MONITOR STATUS*\n\n")
	sb.WriteString(fmt.Sprintf("*Endpoint:* `%s`\n", tn.shopURL))
	sb.WriteString(fmt.Sprintf("*Poll Interval:* `%s`\n", pollInterval))
	sb.WriteString(fmt.Sprintf("*Products:* %d (%d variants tracked)\n", len(products), totalVariants))
	sb.WriteString(fmt.Sprintf("*Available:* %d | *Out of Stock:* %d\n", inStock, outOfStock))

	availabilityRate := 0.0
	if totalVariants > 0 {
		availabilityRate = (float64(inStock) / float64(totalVariants)) * 100
	}
	sb.WriteString(fmt.Sprintf("*Availability Rate:* %.1f%%\n\n", availabilityRate))

	if len(inStockItems) > 0 {
		sb.WriteString("*Currently In Stock:*\n\n")
		for _, item := range inStockItems {
			sb.WriteString(item + "\n\n")
		}
	} else {
		sb.WriteString("_No items currently available_\n\n")
	}

	sb.WriteString("Monitor initialized • ETag caching enabled")

	msg := tgbotapi.NewMessage(tn.chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

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
