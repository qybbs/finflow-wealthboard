package storage

import (
	"os"
	"sync"
)

type StorageManager struct {
	incomeMutex		sync.Mutex
	expenseMutex	sync.Mutex
	portfolioMutex	sync.Mutex
	budgetMutex		sync.Mutex

	incomeFile		string
	expenseFile		string
	portfolioFile	string
	budgetFile		string
}

func NewStorageManager(incomePath, expensePath, portfolioPath, budgetPath string) *StorageManager {
	return &StorageManager{
		incomeFile: 	incomePath,
		expenseFile:	expensePath,
		portfolioFile: 	portfolioPath,
		budgetFile:		budgetPath,
	}
}

func (sm *StorageManager) AddExpense(tx Transaction) error {
	sm.expenseMutex.Lock()

	defer sm.expenseMutex.Unlock()

	txs, err := LoadTransactions(sm.expenseFile)
	if err != nil {
		txs = []Transaction{}
	}

	txs = append(txs, tx)

	return SaveTransactions(sm.expenseFile, txs)
}

func (sm *StorageManager) GetExpenses() ([]Transaction, error) {
	sm.expenseMutex.Lock()
	defer sm.expenseMutex.Unlock()

	txs, err := LoadTransactions(sm.expenseFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Transaction{}, nil
		}

		return nil, err
	}

	return txs, nil
}

func (sm *StorageManager) AddIncome(tx Transaction) error {
	sm.incomeMutex.Lock()

	defer sm.incomeMutex.Unlock()

	txs, err := LoadTransactions(sm.incomeFile)
	if err != nil {
		txs = []Transaction{}
	}

	txs = append(txs, tx)

	return SaveTransactions(sm.incomeFile, txs)
}

func (sm *StorageManager) GetIncomes() ([]Transaction, error) {
	sm.incomeMutex.Lock()
	defer sm.incomeMutex.Unlock()

	txs, err := LoadTransactions(sm.incomeFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Transaction{}, nil
		}

		return nil, err
	}

	return txs, nil
}

func (sm *StorageManager) UpdatePortfolio(port *Portfolio) error {
	sm.portfolioMutex.Lock()

	defer sm.portfolioMutex.Unlock()

	return SavePortfolio(sm.portfolioFile, port)
}

func (sm *StorageManager) GetPortfolio() (*Portfolio, error) {
	sm.portfolioMutex.Lock()
	defer sm.portfolioMutex.Unlock()

	port, err := LoadPortfolio(sm.portfolioFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Portfolio{}, nil
		}

		return nil, err
	}

	return port, nil
}

func (sm *StorageManager) UpdateBudgets(budgets []Budget) error {
	sm.budgetMutex.Lock()
	defer sm.budgetMutex.Unlock()
	return SaveBudgets(sm.budgetFile, budgets)
}

func (sm *StorageManager) GetBudgets() ([]Budget, error) {
	sm.budgetMutex.Lock()
	defer sm.budgetMutex.Unlock()

	budgets, err := LoadBudgets(sm.budgetFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Budget{}, nil
		}
		return nil, err
	}
	return budgets, nil
}

func (sm *StorageManager) BackupAllFiles() error {
	sm.incomeMutex.Lock()
	defer sm.incomeMutex.Unlock()

	sm.expenseMutex.Lock()
	defer sm.expenseMutex.Unlock()

	sm.portfolioMutex.Lock()
	defer sm.portfolioMutex.Unlock()

	sm.budgetMutex.Lock()
	defer sm.budgetMutex.Unlock()
	
	errIncome := BackupData(sm.incomeFile)
	errExpense := BackupData(sm.expenseFile)
	errPortfolio := BackupData(sm.portfolioFile)
	errBudget := BackupData(sm.budgetFile)

	if errIncome != nil { return errIncome }
	if errExpense != nil { return errExpense }
	if errPortfolio != nil { return errPortfolio }
	if errBudget != nil { return errBudget }

	return nil
}