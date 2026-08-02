package handlers

import (
	"encoding/json"
	"finflow-wealthboard/services"
	"finflow-wealthboard/storage"
	"net/http"
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