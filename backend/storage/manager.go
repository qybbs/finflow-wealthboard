package storage

import (
	"database/sql"
)

type StorageManager struct {
	DB *sql.DB
}

func NewStorageManager(db *sql.DB) *StorageManager {
	return &StorageManager{
		DB: db,
	}
}

func (sm *StorageManager) AddExpense(tx Transaction) error {
	_, err := sm.DB.Exec(
		`INSERT INTO transactions (id, date, type, category, amount, description, method) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tx.ID, tx.Date, tx.Type, tx.Category, tx.Amount, tx.Description, tx.Method,
	)
	return err
}

func (sm *StorageManager) GetExpenses() ([]Transaction, error) {
	return sm.getTransactionsByType("EXPENSE")
}

func (sm *StorageManager) AddIncome(tx Transaction) error {
	_, err := sm.DB.Exec(
		`INSERT INTO transactions (id, date, type, category, amount, description, method) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tx.ID, tx.Date, tx.Type, tx.Category, tx.Amount, tx.Description, tx.Method,
	)
	return err
}

func (sm *StorageManager) GetIncomes() ([]Transaction, error) {
	return sm.getTransactionsByType("INCOME")
}

func (sm *StorageManager) getTransactionsByType(txType string) ([]Transaction, error) {
	rows, err := sm.DB.Query(
		`SELECT id, date, type, category, amount, description, method 
		 FROM transactions WHERE type = $1 ORDER BY date DESC`,
		txType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.Date, &tx.Type, &tx.Category, &tx.Amount, &tx.Description, &tx.Method); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (sm *StorageManager) GetPortfolio() (*Portfolio, error) {
	portfolio := &Portfolio{
		Assets:         []Asset{},
		JournalEntries: []JournalEntry{},
	}

	// Load assets
	rows, err := sm.DB.Query(`SELECT id, type, code, quantity, price_per_unit, average_price, current_price FROM assets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.ID, &a.Type, &a.Code, &a.Quantity, &a.PricePerUnit, &a.AveragePrice, &a.CurrentPrice); err != nil {
			return nil, err
		}
		portfolio.Assets = append(portfolio.Assets, a)
	}

	// Load journal entries
	jRows, err := sm.DB.Query(`SELECT id, date, asset_id, transaction_id, type, quantity, price_per_unit, fee, total FROM journal_entries ORDER BY date DESC`)
	if err != nil {
		return nil, err
	}
	defer jRows.Close()

	for jRows.Next() {
		var j JournalEntry
		var txID sql.NullString
		if err := jRows.Scan(&j.ID, &j.Date, &j.AssetID, &txID, &j.Type, &j.Quantity, &j.PricePerUnit, &j.Fee, &j.Total); err != nil {
			return nil, err
		}
		if txID.Valid {
			j.TransactionID = txID.String
		}
		portfolio.JournalEntries = append(portfolio.JournalEntries, j)
	}

	return portfolio, nil
}

func (sm *StorageManager) UpdatePortfolio(port *Portfolio) error {
	// Not strictly needed in a relational DB to update the whole portfolio at once
	// if we're updating assets individually, but we'll provide a full sync for backward compatibility if needed.
	tx, err := sm.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, asset := range port.Assets {
		_, err := tx.Exec(
			`INSERT INTO assets (id, type, code, quantity, price_per_unit, average_price, current_price) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7) 
			 ON CONFLICT (id) DO UPDATE SET 
			 quantity = EXCLUDED.quantity, 
			 price_per_unit = EXCLUDED.price_per_unit, 
			 average_price = EXCLUDED.average_price, 
			 current_price = EXCLUDED.current_price`,
			asset.ID, asset.Type, asset.Code, asset.Quantity, asset.PricePerUnit, asset.AveragePrice, asset.CurrentPrice,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (sm *StorageManager) UpdateBudgets(budgets []Budget) error {
	tx, err := sm.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, budget := range budgets {
		_, err := tx.Exec(
			`INSERT INTO budgets (category, limit_amount, interval) 
			 VALUES ($1, $2, $3) 
			 ON CONFLICT (category) DO UPDATE SET 
			 limit_amount = EXCLUDED.limit_amount, 
			 interval = EXCLUDED.interval`,
			budget.Category, budget.Limit, budget.Interval,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (sm *StorageManager) GetBudgets() ([]Budget, error) {
	rows, err := sm.DB.Query(`SELECT category, limit_amount, interval FROM budgets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.Category, &b.Limit, &b.Interval); err != nil {
			return nil, err
		}
		budgets = append(budgets, b)
	}
	return budgets, nil
}

func (sm *StorageManager) UpdateTransaction(tx Transaction) error {
	_, err := sm.DB.Exec(
		`UPDATE transactions SET date = $1, type = $2, category = $3, amount = $4, description = $5, method = $6 WHERE id = $7`,
		tx.Date, tx.Type, tx.Category, tx.Amount, tx.Description, tx.Method, tx.ID,
	)
	return err
}

func (sm *StorageManager) DeleteTransaction(id string) error {
	_, err := sm.DB.Exec(`DELETE FROM transactions WHERE id = $1`, id)
	return err
}



type PortfolioTransactionReq struct {
	ID            string  `json:"id"`
	Date          string  `json:"date"`
	AssetID       string  `json:"asset_id"`
	AssetType     string  `json:"asset_type"`
	AssetCode     string  `json:"asset_code"`
	Type          string  `json:"type"` // "BUY", "SELL", "DIVIDEND"
	Quantity      float64 `json:"quantity"`
	PricePerUnit  float64 `json:"price_per_unit"`
	Fee           float64 `json:"fee"`
	Method        string  `json:"method"`
}

func (sm *StorageManager) GetAssetByID(id string) (*Asset, error) {
	var asset Asset
	err := sm.DB.QueryRow(`SELECT id, type, code, quantity, price_per_unit, average_price, current_price FROM assets WHERE id = $1`, id).
		Scan(&asset.ID, &asset.Type, &asset.Code, &asset.Quantity, &asset.PricePerUnit, &asset.AveragePrice, &asset.CurrentPrice)
	if err != nil {
		return nil, err
	}
	// Calculate totals
	asset.TotalValue = asset.Quantity * asset.CurrentPrice
	asset.ProfitLoss = asset.TotalValue - (asset.AveragePrice * asset.Quantity)
	if asset.AveragePrice > 0 && asset.Quantity > 0 {
		asset.ProfitLossPct = (asset.ProfitLoss / (asset.AveragePrice * asset.Quantity)) * 100
	}
	return &asset, nil
}

func (sm *StorageManager) AddPortfolioTransaction(req PortfolioTransactionReq) error {
	tx, err := sm.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure asset exists
	var oldQty, oldAvgPrice float64
	err = tx.QueryRow(`SELECT quantity, average_price FROM assets WHERE id = $1`, req.AssetID).Scan(&oldQty, &oldAvgPrice)
	if err == sql.ErrNoRows {
		// Create new asset
		_, err = tx.Exec(
			`INSERT INTO assets (id, type, code, quantity, price_per_unit, average_price, current_price) 
			 VALUES ($1, $2, $3, 0, 0, 0, 0)`,
			req.AssetID, req.AssetType, req.AssetCode,
		)
		if err != nil { return err }
		oldQty, oldAvgPrice = 0, 0
	} else if err != nil {
		return err
	}

	totalValue := (req.Quantity * req.PricePerUnit)

	// Update asset and cashflow based on type
	if req.Type == "BUY" {
		totalCost := totalValue + req.Fee
		newQty := oldQty + req.Quantity
		var newAvgPrice float64
		if newQty > 0 {
			newAvgPrice = ((oldAvgPrice * oldQty) + totalCost) / newQty
		}
		
		_, err = tx.Exec(`UPDATE assets SET quantity = $1, average_price = $2 WHERE id = $3`, newQty, newAvgPrice, req.AssetID)
		if err != nil { return err }

		// Add expense
		_, err = tx.Exec(
			`INSERT INTO transactions (id, date, type, category, amount, description, method) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			"tx_buy_" + req.ID, req.Date, "EXPENSE", "Investasi", totalCost, "Beli " + req.AssetCode, req.Method,
		)
		if err != nil { return err }

	} else if req.Type == "SELL" {
		totalProceeds := totalValue - req.Fee
		newQty := oldQty - req.Quantity
		if newQty < 0 { newQty = 0 } // safeguard
		newAvgPrice := oldAvgPrice
		if newQty == 0 { newAvgPrice = 0 }

		_, err = tx.Exec(`UPDATE assets SET quantity = $1, average_price = $2 WHERE id = $3`, newQty, newAvgPrice, req.AssetID)
		if err != nil { return err }

		// Add income
		_, err = tx.Exec(
			`INSERT INTO transactions (id, date, type, category, amount, description, method) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			"tx_sell_" + req.ID, req.Date, "INCOME", "Divestasi", totalProceeds, "Jual " + req.AssetCode, req.Method,
		)
		if err != nil { return err }

	} else if req.Type == "DIVIDEND" {
		totalProceeds := totalValue - req.Fee
		// Add income
		_, err = tx.Exec(
			`INSERT INTO transactions (id, date, type, category, amount, description, method) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			"tx_div_" + req.ID, req.Date, "INCOME", "Dividends/Interest", totalProceeds, "Dividen/Kupon " + req.AssetCode, req.Method,
		)
		if err != nil { return err }
	}

	// Insert journal entry
	txTotal := totalValue
	if req.Type == "BUY" { txTotal += req.Fee } else { txTotal -= req.Fee }
	
	_, err = tx.Exec(
		`INSERT INTO journal_entries (id, date, asset_id, transaction_id, type, quantity, price_per_unit, fee, total) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		req.ID, req.Date, req.AssetID, nil, req.Type, req.Quantity, req.PricePerUnit, req.Fee, txTotal,
	)
	if err != nil { return err }

	return tx.Commit()
}

func (sm *StorageManager) UpdateAssetPrice(assetID string, price float64) error {
	_, err := sm.DB.Exec(`UPDATE assets SET current_price = $1 WHERE id = $2`, price, assetID)
	return err
}