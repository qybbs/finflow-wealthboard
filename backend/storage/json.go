package storage

import (
	"encoding/json"
	"os"
)

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

func SavePortfolio(filepath string, portfolio *Portfolio) error {
	data, err := json.MarshalIndent(portfolio, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}