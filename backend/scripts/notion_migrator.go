package main

import (
	"fmt"
	"log"
	"os"
)

// Script migrasi mandiri dari Notion ke lokal.
// Bisa dijalankan dengan: go run scripts/notion_migrator.go

const (
	NotionAPIURL = "https://api.notion.com/v1/databases/%s/query"
)

func main() {
	fmt.Println("🚀 Memulai Migrasi Data dari Notion...")
	apiKey := os.Getenv("NOTION_API_KEY")
	dbIncome := os.Getenv("NOTION_INCOME_DB_ID")
	dbExpense := os.Getenv("NOTION_EXPENSE_DB_ID")
	dbPortfolio := os.Getenv("NOTION_PORTFOLIO_DB_ID")

	if apiKey == "" {
		fmt.Println("Peringatan: NOTION_API_KEY tidak ditemukan di environment variables.")
		fmt.Println("Migrasi ini hanya berjalan sebagai simulasi dengan data tiruan (Mock) agar UI bisa dicoba.")
		generateMockData()
		return
	}

	if dbIncome == "" || dbExpense == "" || dbPortfolio == "" {
		log.Fatalf("Harap set NOTION_INCOME_DB_ID, NOTION_EXPENSE_DB_ID, dan NOTION_PORTFOLIO_DB_ID")
	}

	fmt.Println("Menghubungkan ke Notion API...")
	// Logika pemanggilan HTTP ke Notion akan dieksekusi di sini jika API Key tersedia
	// ...
}

// generateMockData membuat dummy data (pembersihan/pembulatan unit saham)
func generateMockData() {
	incomeData := `id,date,type,category,amount,description,method
tx_inc_1,2026-08-01,INCOME,Gaji,15000000,Gaji Bulanan Agustus,BANK
`
	expenseData := `id,date,type,category,amount,description,method
tx_exp_1,2026-08-02,EXPENSE,Makanan,50000,Makan Siang Nasi Padang,CASH
tx_exp_2,2026-08-03,EXPENSE,Transportasi,150000,Bensin Mobil,EWALLET
tx_exp_3,2026-08-04,EXPENSE,Makanan,75000,Makan Malam,CASH
tx_exp_4,2026-08-05,EXPENSE,Hiburan,200000,Nonton Bioskop,BANK
`
	
	portfolioData := `{
  "assets": [
    {
      "id": "asset_1",
      "type": "SAHAM",
      "code": "BBCA",
      "quantity": 1000,
      "price_per_unit": 0,
      "average_price": 9500,
      "current_price": 0,
      "total_value": 0,
      "profit_loss": 0,
      "profit_loss_pct": 0
    },
    {
      "id": "asset_2",
      "type": "SAHAM",
      "code": "BBRI",
      "quantity": 500,
      "price_per_unit": 0,
      "average_price": 4800,
      "current_price": 0,
      "total_value": 0,
      "profit_loss": 0,
      "profit_loss_pct": 0
    },
    {
      "id": "asset_3",
      "type": "EMAS",
      "code": "ANTAM",
      "quantity": 10,
      "price_per_unit": 0,
      "average_price": 1050000,
      "current_price": 0,
      "total_value": 0,
      "profit_loss": 0,
      "profit_loss_pct": 0
    },
    {
      "id": "asset_4",
      "type": "REKSA_DANA",
      "code": "SUCORINVEST_MONEY_MARKET",
      "quantity": 5000,
      "price_per_unit": 1250,
      "average_price": 1200,
      "current_price": 1250,
      "total_value": 0,
      "profit_loss": 0,
      "profit_loss_pct": 0
    }
  ],
  "journal_entries": []
}`

	os.MkdirAll("../data", 0755)
	os.WriteFile("../data/income.csv", []byte(incomeData), 0644)
	os.WriteFile("../data/expense.csv", []byte(expenseData), 0644)
	os.WriteFile("../data/portfolio.json", []byte(portfolioData), 0644)

	fmt.Println("✅ Data migrasi simulasi berhasil ditulis ke folder data/")
}
