package store

import (
	"go-banking/models"
	"sync"
)

// Store holds in-memory data for customers, accounts, and transactions
type Store struct {
	customers    map[string]models.Customer
	accounts     map[string]models.Account
	transactions map[string]models.Transaction
	mutex        sync.RWMutex
}

// GlobalStore is a singleton instance of Store
var GlobalStore = &Store{
	customers:    make(map[string]models.Customer),
	accounts:     make(map[string]models.Account),
	transactions: make(map[string]models.Transaction),
}

// AddCustomer adds a customer to the store
func (s *Store) AddCustomer(customer models.Customer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.customers[customer.ID] = customer
}

// GetCustomerByID retrieves a customer by ID
func (s *Store) GetCustomerByID(id string) (models.Customer, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	customer, exists := s.customers[id]
	return customer, exists
}

// AddAccount adds an account to the store
func (s *Store) AddAccount(account models.Account) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.accounts[account.ID] = account
}

// GetAccountByID retrieves an account by ID
func (s *Store) GetAccountByID(id string) (models.Account, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	account, exists := s.accounts[id]
	return account, exists
}

// UpdateAccount updates an existing account
func (s *Store) UpdateAccount(account models.Account) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.accounts[account.ID] = account
}

// GetAccountsByCustomerID retrieves all accounts for a customer
func (s *Store) GetAccountsByCustomerID(customerID string) []models.Account {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	var accounts []models.Account
	for _, account := range s.accounts {
		if account.CustomerID == customerID {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

// AddTransaction adds a transaction to the store
func (s *Store) AddTransaction(transaction models.Transaction) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.transactions[transaction.ID] = transaction
}

// GetTransactionsByCustomerID retrieves all transactions for a customer's accounts
func (s *Store) GetTransactionsByCustomerID(customerID string) []models.Transaction {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	var transactions []models.Transaction
	// Get all accounts for the customer
	accountIDs := make(map[string]bool)
	for _, account := range s.accounts {
		if account.CustomerID == customerID {
			accountIDs[account.ID] = true
		}
	}
	// Collect transactions for those accounts
	for _, transaction := range s.transactions {
		if accountIDs[transaction.AccountID] {
			transactions = append(transactions, transaction)
		}
	}
	return transactions
}
