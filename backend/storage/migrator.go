package storage

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"strconv"
)

func LoadTransactions(filepath string) ([]Transaction, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, _ = reader.Read() // skip header

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var transactions []Transaction
	for i, record := range records {
		if len(record) < 7 {
			log.Printf("LoadTransactions: skipping malformed row %d (expected 7 columns, got %d)", i+2, len(record))
			continue
		}
		amount, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			log.Printf("LoadTransactions: skipping row %d: invalid amount %q: %v", i+2, record[4], err)
			continue
		}
		transactions = append(transactions, Transaction{
			ID:          record[0],
			Date:        record[1],
			Type:        record[2],
			Category:    record[3],
			Amount:      amount,
			Description: record[5],
			Method:      record[6],
		})
	}

	return transactions, nil
}

func LoadPortfolio(filepath string) (*Portfolio, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var portfolio Portfolio
	err = json.Unmarshal(data, &portfolio)
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

func LoadBudgets(filepath string) ([]Budget, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var budgets []Budget
	err = json.Unmarshal(data, &budgets)
	if err != nil {
		return nil, err
	}
	return budgets, nil
}

func MigrateData(db *sql.DB, incomePath, expensePath, portfolioPath, budgetPath string) error {
	migrated := false

	if err := migrateTransactions(db, incomePath, "INCOME"); err != nil {
		log.Printf("Failed to migrate incomes: %v\n", err)
	} else {
		migrated = true
	}

	if err := migrateTransactions(db, expensePath, "EXPENSE"); err != nil {
		log.Printf("Failed to migrate expenses: %v\n", err)
	} else {
		migrated = true
	}

	if err := migratePortfolio(db, portfolioPath); err != nil {
		log.Printf("Failed to migrate portfolio: %v\n", err)
	} else {
		migrated = true
	}
	
	if err := migrateBudgets(db, budgetPath); err != nil {
		log.Printf("Failed to migrate budgets: %v\n", err)
	} else {
		migrated = true
	}

	if migrated {
		log.Println("Data migration completed.")
	}

	return nil
}

func migrateTransactions(db *sql.DB, filepath string, txType string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to migrate
	}

	txs, err := LoadTransactions(filepath)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		_, err := db.Exec(
			`INSERT INTO transactions (id, date, type, category, amount, description, method) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) DO NOTHING`,
			tx.ID, tx.Date, txType, tx.Category, tx.Amount, tx.Description, tx.Method,
		)
		if err != nil {
			log.Printf("Error inserting transaction %s: %v", tx.ID, err)
		}
	}

	return os.Rename(filepath, filepath+".bak")
}

func migratePortfolio(db *sql.DB, filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil
	}

	portfolio, err := LoadPortfolio(filepath)
	if err != nil {
		return err
	}

	for _, asset := range portfolio.Assets {
		_, err := db.Exec(
			`INSERT INTO assets (id, type, code, quantity, price_per_unit, average_price, current_price)
			 VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) DO NOTHING`,
			asset.ID, asset.Type, asset.Code, asset.Quantity, asset.PricePerUnit, asset.AveragePrice, asset.CurrentPrice,
		)
		if err != nil {
			log.Printf("Error inserting asset %s: %v", asset.ID, err)
		}
	}

	for _, entry := range portfolio.JournalEntries {
		_, err := db.Exec(
			`INSERT INTO journal_entries (id, date, asset_id, transaction_id, type, quantity, price_per_unit, fee, total)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (id) DO NOTHING`,
			entry.ID, entry.Date, entry.AssetID, entry.TransactionID, entry.Type, entry.Quantity, entry.PricePerUnit, entry.Fee, entry.Total,
		)
		if err != nil {
			log.Printf("Error inserting journal entry %s: %v", entry.ID, err)
		}
	}

	return os.Rename(filepath, filepath+".bak")
}

func migrateBudgets(db *sql.DB, filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil
	}

	budgets, err := LoadBudgets(filepath)
	if err != nil {
		return err
	}

	for _, budget := range budgets {
		_, err := db.Exec(
			`INSERT INTO budgets (category, limit_amount, interval)
			 VALUES ($1, $2, $3) ON CONFLICT (category) DO NOTHING`,
			budget.Category, budget.Limit, budget.Interval,
		)
		if err != nil {
			log.Printf("Error inserting budget %s: %v", budget.Category, err)
		}
	}

	return os.Rename(filepath, filepath+".bak")
}
