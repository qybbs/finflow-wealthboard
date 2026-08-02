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

	// Dalam skenario asli, h.Store.GetIncome() akan dipanggil.
	// Karena kita menggunakan mock data untuk saat ini, kita lakukan perhitungan sederhana:
	expenses, _ := h.Store.GetExpenses()
	
	// Hitung total pengeluaran
	totalExp := 0.0
	for _, exp := range expenses {
		totalExp += exp.Amount
	}

	// Contoh statis Pemasukan & Kas (Ideanya di-fetch dari income.csv)
	totalInc := 15000000.0 // 15 Juta
	totalCash := 50000000.0 // 50 Juta di kas

	savingsRate := 0.0
	if totalInc > 0 {
		savingsRate = ((totalInc - totalExp) / totalInc) * 100
	}

	avgMonthlyExp := totalExp * 4 // Misalkan totalExp adalah seminggu
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
