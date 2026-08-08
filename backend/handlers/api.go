package handlers

import (
	"encoding/json"
	"finflow-wealthboard/services"
	"finflow-wealthboard/storage"
	"net/http"
	"time"
)

type APIHandler struct {
	Store    *storage.StorageManager
	StockSvc *services.StockService
	GoldSvc  *services.GoldService
}

func (h *APIHandler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	expenses, err := h.Store.GetExpenses()
	if err != nil {
		http.Error(w, "Gagal membaca data: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expenses)
}

func (h *APIHandler) GetIncomes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	incomes, err := h.Store.GetIncomes()
	if err != nil {
		http.Error(w, "Gagal membaca data: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(incomes)
}

func (h *APIHandler) AddExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var newTx storage.Transaction
	err := json.NewDecoder(r.Body).Decode(&newTx)
	if err != nil {
		http.Error(w, "Data JSON tidak valid", http.StatusBadRequest)
		return
	}

	err = h.Store.AddExpense(newTx)
	if err != nil {
		http.Error(w, "Gagal menyimpan data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success", "message": "Pengeluaran berhasil ditambahkan"}`))


}

func (h *APIHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	portfolio, err := h.Store.GetPortfolio()
	if err != nil {
		http.Error(w, "Gagal membaca data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update portfolio with dynamic prices
	for i, asset := range portfolio.Assets {
		currentPrice := asset.CurrentPrice
		if asset.Type == "SAHAM" {
			if price, err := h.StockSvc.GetStockPrice(asset.Code); err == nil {
				currentPrice = price
			}
		} else if asset.Type == "EMAS" {
			if price, err := h.GoldSvc.GetGoldPrice(); err == nil {
				currentPrice = price
			}
		} // REKSA_DANA keeps its manual price

		// Recalculate values
		portfolio.Assets[i].CurrentPrice = currentPrice
		portfolio.Assets[i].TotalValue = asset.Quantity * currentPrice
		portfolio.Assets[i].ProfitLoss = portfolio.Assets[i].TotalValue - (asset.AveragePrice * asset.Quantity)
		if asset.AveragePrice > 0 && asset.Quantity > 0 {
			portfolio.Assets[i].ProfitLossPct = (portfolio.Assets[i].ProfitLoss / (asset.AveragePrice * asset.Quantity)) * 100
		} else {
			portfolio.Assets[i].ProfitLossPct = 0
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(portfolio)
}

func (h *APIHandler) UpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var portfolio storage.Portfolio
	err := json.NewDecoder(r.Body).Decode(&portfolio)
	if err != nil {
		http.Error(w, "Data JSON tidak valid", http.StatusBadRequest)
		return
	}

	err = h.Store.UpdatePortfolio(&portfolio)
	if err != nil {
		http.Error(w, "Gagal menyimpan data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success", "message": "Portfolio berhasil di-update"}`))


}

type AccountBalance struct {
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

func (h *APIHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	incomes, _ := h.Store.GetIncomes()
	expenses, _ := h.Store.GetExpenses()

	balances := make(map[string]float64)

	for _, inc := range incomes {
		balances[inc.Method] += inc.Amount
	}

	for _, exp := range expenses {
		balances[exp.Method] -= exp.Amount
	}

	var accounts []AccountBalance
	for name, bal := range balances {
		if name != "" {
			accounts = append(accounts, AccountBalance{Name: name, Balance: bal})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accounts)
}

type BudgetResponse struct {
	Category  string  `json:"category"`
	Limit     float64 `json:"limit"`
	Interval  string  `json:"interval"`
	Spent     float64 `json:"spent"`
	Remaining float64 `json:"remaining"`
}

func (h *APIHandler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	budgets, _ := h.Store.GetBudgets()
	expenses, _ := h.Store.GetExpenses()

	now := time.Now()
	var response []BudgetResponse

	for _, b := range budgets {
		spent := 0.0
		for _, exp := range expenses {
			if exp.Category == b.Category {
				expDate, err := time.Parse("2006-01-02", exp.Date)
				if err == nil {
					if b.Interval == "WEEKLY" {
						if now.Sub(expDate).Hours() <= 24*7 && expDate.Before(now.Add(24*time.Hour)) {
							spent += exp.Amount
						}
					} else { // MONTHLY
						if expDate.Month() == now.Month() && expDate.Year() == now.Year() {
							spent += exp.Amount
						}
					}
				}
			}
		}
		
		response = append(response, BudgetResponse{
			Category:  b.Category,
			Limit:     b.Limit,
			Interval:  b.Interval,
			Spent:     spent,
			Remaining: b.Limit - spent,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) UpdateBudgets(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var budgets []storage.Budget
	err := json.NewDecoder(r.Body).Decode(&budgets)
	if err != nil {
		http.Error(w, "Data JSON tidak valid", http.StatusBadRequest)
		return
	}

	err = h.Store.UpdateBudgets(budgets)
	if err != nil {
		http.Error(w, "Gagal menyimpan data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success", "message": "Budget berhasil di-update"}`))
}

func (h *APIHandler) AddIncome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var newTx storage.Transaction
	err := json.NewDecoder(r.Body).Decode(&newTx)
	if err != nil {
		http.Error(w, "Data JSON tidak valid", http.StatusBadRequest)
		return
	}

	err = h.Store.AddIncome(newTx)
	if err != nil {
		http.Error(w, "Gagal menyimpan data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success", "message": "Pemasukan berhasil ditambahkan"}`))
}

func (h *APIHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var updatedTx storage.Transaction
	err := json.NewDecoder(r.Body).Decode(&updatedTx)
	if err != nil {
		http.Error(w, "Data JSON tidak valid", http.StatusBadRequest)
		return
	}

	err = h.Store.UpdateTransaction(updatedTx)
	if err != nil {
		http.Error(w, "Gagal memperbarui data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success", "message": "Transaksi berhasil diperbarui"}`))
}

func (h *APIHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ID == "" {
		http.Error(w, "Data JSON tidak valid atau ID kosong", http.StatusBadRequest)
		return
	}

	err = h.Store.DeleteTransaction(req.ID)
	if err != nil {
		http.Error(w, "Gagal menghapus data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success", "message": "Transaksi berhasil dihapus"}`))
}

func (h *APIHandler) AddPortfolioTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req storage.PortfolioTransactionReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Data JSON tidak valid", http.StatusBadRequest)
		return
	}

	err = h.Store.AddPortfolioTransaction(req)
	if err != nil {
		http.Error(w, "Gagal menyimpan transaksi portofolio: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success", "message": "Transaksi portofolio berhasil dicatat"}`))
}

func (h *APIHandler) UpdateAssetPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AssetID string  `json:"asset_id"`
		Price   float64 `json:"price"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Data JSON tidak valid", http.StatusBadRequest)
		return
	}

	err = h.Store.UpdateAssetPrice(req.AssetID, req.Price)
	if err != nil {
		http.Error(w, "Gagal memperbarui harga: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success", "message": "Harga aset berhasil diperbarui"}`))
}