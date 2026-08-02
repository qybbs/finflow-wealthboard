package storage

type Transaction struct {
	ID          string  `json:"id"`
	Date        string  `json:"date"` // YYYY-MM-DD
	Type        string  `json:"type"` // "INCOME" | "EXPENSE" | "TRANSFER"
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Method      string  `json:"method"` // "CASH" | "BANK" | "EWALLET"
}

type Asset struct {
	ID             string  `json:"id"`
	Type           string  `json:"type"` // "SAHAM" | "REKSA_DANA" | "EMAS"
	Code           string  `json:"code"` // Ticker (e.g., BBCA, GOTO)
	Quantity       float64 `json:"quantity"` // Jumlah lembar/unit
	PricePerUnit   float64 `json:"price_per_unit"` // Harga per lembar/unit
	AveragePrice   float64 `json:"average_price"` // Harga beli rata-rata
	CurrentPrice   float64 `json:"current_price"` // Harga pasar terkini (ditarik dari API)
	TotalValue     float64 `json:"total_value"` // Quantity * CurrentPrice
	ProfitLoss     float64 `json:"profit_loss"` // TotalValue - (AveragePrice * Quantity)
	ProfitLossPct  float64 `json:"profit_loss_pct"`
}

type JournalEntry struct {
	ID        string  `json:"id"`
	Date      string  `json:"date"` // YYYY-MM-DD
	AssetID   int     `json:"asset_id"` // Relasi ke tabel Asset
	TransactionID int `json:"transaction_id"` // Relasi ke tabel Transaction
	Type      string  `json:"type"` // "BUY" | "SELL" | "DIVIDEND" | "ADJUSTMENT"
	Quantity  float64 `json:"quantity"` // Positif untuk buy, Negatif untuk sell
	PricePerUnit float64 `json:"price_per_unit"`
	Fee       float64 `json:"fee"`
	Total     float64 `json:"total"` // (Quantity * PricePerUnit) + Fee
}

type Portfolio struct {
	Assets			[]Asset `json:"assets"`
	JournalEntries 	[]JournalEntry `json:"journal_entries"`
}