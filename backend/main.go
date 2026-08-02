package main

import (
	"finflow-wealthboard/handlers"
	"finflow-wealthboard/services"
	"finflow-wealthboard/storage"
	"fmt"
	"log"
	"net/http"
)

func main() {
	store := storage.NewStorageManager(
		"../data/income.csv",
		"../data/expense.csv",
		"../data/portfolio.json",
		"../data/budget.json",
	)

	fmt.Println("Memulai FinFlow Wealthboard...")
	if err := store.BackupAllFiles(); err != nil {
		log.Printf("Peringatan: Gagal melakukan backup: %v\n", err)
	} else {
		fmt.Println("Auto-backup selesai (atau file belum ada)")
	}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("../static"))
	mux.Handle("/", fileServer)

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "API FinFlow Wealthboard berjalan lancar!"}`))
	})

	stockSvc := services.NewStockService()
	goldSvc := services.NewGoldService()

	api := &handlers.APIHandler{
		Store:    store,
		StockSvc: stockSvc,
		GoldSvc:  goldSvc,
	}

	mux.HandleFunc("/api/expenses", api.GetExpenses)
	mux.HandleFunc("/api/incomes", api.GetIncomes)
	mux.HandleFunc("/api/expenses/add", api.AddExpense)
	mux.HandleFunc("/api/portfolio", api.GetPortfolio)
	mux.HandleFunc("/api/portfolio/update", api.UpdatePortfolio)
	mux.HandleFunc("/api/analytics", api.GetAnalytics)
	mux.HandleFunc("/api/accounts", api.GetAccounts)
	mux.HandleFunc("/api/budgets", api.GetBudgets)
	mux.HandleFunc("/api/budgets/update", api.UpdateBudgets)

	port := ":8080"
	fmt.Printf("Server berhasil berjalan di http://localhost%s\n", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server berhenti karena error: %v", err)
	}
}