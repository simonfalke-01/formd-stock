package main

// ShopifyResponse represents the collection products API response
type ShopifyResponse struct {
	Products []Product `json:"products"`
}

// Product represents a Shopify product
type Product struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Handle    string    `json:"handle"`
	Available bool      `json:"available"`
	Variants  []Variant `json:"variants"`
	Tags      []string  `json:"tags"`
}

// Variant represents a product variant
type Variant struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Available bool   `json:"available"`
	Price     string `json:"price"`
	SKU       string `json:"sku"`
}

// StockChange represents a detected availability change
type StockChange struct {
	ProductID     int64
	ProductTitle  string
	ProductHandle string
	VariantID     int64
	VariantTitle  string
	VariantPrice  string
	VariantSKU    string
	ProductURL    string
	WasAvailable  bool
	IsAvailable   bool
}

// IsNewStock returns true if this is a new stock (became available)
func (sc StockChange) IsNewStock() bool {
	return !sc.WasAvailable && sc.IsAvailable
}
