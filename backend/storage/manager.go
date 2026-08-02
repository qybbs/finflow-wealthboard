package storage

import (
	"sync"
)

type StorageManager struct {
	incomeMutex		sync.Mutex
	expenseMutex	sync.Mutex
	portfolioMutex	sync.Mutex

	incomeFile		string
	expenseFile		string
	portfolioFile	string
}

func NewStorageManager(incomePath, expensePath, portfolioPath string) *StorageManager {
	return &StorageManager{
		incomeFile: 	incomePath,
		expenseFile:	expensePath,
		portfolioFile: 	portfolioPath,
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

	return LoadTransactions(sm.expenseFile)
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

	return LoadTransactions(sm.incomeFile)
}

func (sm *StorageManager) UpdatePortfolio(port *Portfolio) error {
	sm.portfolioMutex.Lock()

	defer sm.portfolioMutex.Unlock()

	return SavePortfolio(sm.portfolioFile, port)
}

func (sm *StorageManager) GetPortfolio() (*Portfolio, error) {
	sm.portfolioMutex.Lock()
	defer sm.portfolioMutex.Unlock()

	return LoadPortfolio(sm.portfolioFile)
}

func (sm *StorageManager) BackupAllFiles() error {
	sm.incomeMutex.Lock()
	defer sm.incomeMutex.Unlock()

	sm.expenseMutex.Lock()
	defer sm.expenseMutex.Unlock()

	sm.portfolioMutex.Lock()
	defer sm.portfolioMutex.Unlock()
	
	errIncome := BackupData(sm.incomeFile)
	errExpense := BackupData(sm.expenseFile)
	errPortfolio := BackupData(sm.portfolioFile)

	if errIncome != nil {
		return errIncome
	}

	if errExpense != nil {
		return errExpense
	}

	if errPortfolio != nil {
		return errPortfolio
	}

	return nil
}