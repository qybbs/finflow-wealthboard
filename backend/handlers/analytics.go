package handlers

import (
	"encoding/json"
	"net/http"
)

// AnalyticsResponse berisi data evaluasi keuangan
type AnalyticsResponse struct {
	SavingsRate      float64 `json:"savings_rate"`
	EmergencyRunRate float64 `json:"emergency_run_rate"`
	TotalIncome      float64 `json:"total_income"`
	TotalExpenses    float64 `json:"total_expenses"`
	TotalCash        float64 `json:"total_cash"`
	Recommendation   string  `json:"recommendation"`
}

func (h *APIHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get Total Expenses
	var totalExp float64
	err := h.Store.DB.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE type = 'EXPENSE'`).Scan(&totalExp)
	if err != nil {
		totalExp = 0
	}

	// Get Total Income
	var totalInc float64
	err = h.Store.DB.QueryRow(`SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE type = 'INCOME'`).Scan(&totalInc)
	if err != nil {
		totalInc = 0
	}

	totalCash = totalInc - totalExp

	savingsRate := 0.0
	if totalInc > 0 {
		savingsRate = ((totalInc - totalExp) / totalInc) * 100
	}

	// Get number of unique months in expenses to calculate avgMonthlyExp
	var uniqueMonths int
	err = h.Store.DB.QueryRow(`SELECT COUNT(DISTINCT SUBSTR(date, 1, 7)) FROM transactions WHERE type = 'EXPENSE'`).Scan(&uniqueMonths)
	if err != nil || uniqueMonths == 0 {
		uniqueMonths = 1
	}

	avgMonthlyExp := totalExp / float64(uniqueMonths)
	emergencyRunRate := 0.0
	if avgMonthlyExp > 0 {
		emergencyRunRate = totalCash / avgMonthlyExp
	}

	// Rekomendasi Asisten
	recommendation := "Keuangan Anda terlihat sehat! "
	if savingsRate < 20 {
		recommendation += "Namun Savings Rate Anda di bawah 20%. Pertimbangkan untuk menabung lebih banyak. "
	}
	if emergencyRunRate < 6 {
		recommendation += "Dana darurat Anda kurang dari 6 bulan pengeluaran. Tingkatkan dana darurat Anda segera."
	}

	response := AnalyticsResponse{
		SavingsRate:      savingsRate,
		EmergencyRunRate: emergencyRunRate,
		TotalIncome:      totalInc,
		TotalExpenses:    totalExp,
		TotalCash:        totalCash,
		Recommendation:   recommendation,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
