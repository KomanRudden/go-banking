package models

import "time"

// Customer represents a bank customer
type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Account represents a bank account (Current or Savings)
type Account struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customerId"`
	Type       string    `json:"type"` // "current" or "savings"
	Balance    float64   `json:"balance"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID            string    `json:"id"`
	AccountID     string    `json:"accountId"`
	CustomerID    string    `json:"customerId"`
	Type          string    `json:"type"` // "transfer", "payment", "bonus", "interest", "fee"
	Amount        float64   `json:"amount"`
	FromAccountID string    `json:"fromAccountId,omitempty"`
	ToAccountID   string    `json:"toAccountId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}
