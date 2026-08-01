package storage

import (
	"encoding/csv"
	"fmt"
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

	_, _ = reader.Read()

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var transactions []Transaction

	for _, record := range records {
		amount, _ := strconv.ParseFloat(record[4], 64)

		tx := Transaction{
			ID:		record[0],
			Date:	record[1],
			Type: record[2],
			Category: record[3],
			Amount: amount,
			Description: record[5],
			Method: record[6],
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func SaveTransactions(filepath string, txs []Transaction) error {
	records := [][]string{{"ID", "Date", "Type", "Category", "Amount", "Description", "Method"}}
	
	for _, tx := range txs {
		records = append(records, []string{
			tx.ID,
			tx.Date,
			tx.Type,
			tx.Category,
			fmt.Sprintf("%.2f", tx.Amount),
			tx.Description,
			tx.Method,
		})
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	return writer.WriteAll(records)
}