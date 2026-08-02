package handlers

import (
	"encoding/json"
	"finflow-wealthboard/storage"
	"net/http"
)

type APIHandler struct {
	Store *storage.StorageManager
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
		http.Error(w, "Gagal membaca data: " + err.Error(), http.StatusInternalServerError)
		return
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