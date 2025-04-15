package bankz

import (
	"errors"
	"fmt"
	"time"
)

// MockAccount represents a Bank Z account
type MockAccount struct {
	AccountID string
	Balance   float64
}

// Token represents an OAuth access token
type Token struct {
	AccessToken string
	ExpiresAt   time.Time
}

// MockClient simulates Bank Z's API with OAuth
type MockClient struct {
	accounts      map[string]MockAccount
	clientID      string
	clientSecret  string
	token         *Token
	tokenEndpoint string // Mock endpoint for token requests
}

// NewMockClient initializes the client with mock credentials
func NewMockClient() *MockClient {
	return &MockClient{
		accounts: map[string]MockAccount{
			"bankz-acc-123": {AccountID: "bankz-acc-123", Balance: 1000.0},
			"bankz-acc-456": {AccountID: "bankz-acc-456", Balance: 500.0},
		},
		clientID:      "mock-client-id",
		clientSecret:  "mock-client-secret",
		tokenEndpoint: "https://mock.bankz.com/oauth/token",
	}
}

// GetToken simulates requesting an OAuth token
func (c *MockClient) GetToken() (string, error) {
	// Check if token exists and is valid
	if c.token != nil && c.token.ExpiresAt.After(time.Now()) {
		return c.token.AccessToken, nil
	}

	// Simulate token request
	if c.clientID != "mock-client-id" || c.clientSecret != "mock-client-secret" {
		return "", errors.New("invalid client credentials")
	}

	// Generate a new mock token
	tokenID := fmt.Sprintf("token-%d", time.Now().UnixNano())
	c.token = &Token{
		AccessToken: tokenID,
		ExpiresAt:   time.Now().Add(1 * time.Hour), // Token valid for 1 hour
	}
	fmt.Printf("Generated new Bank Z token: %s\n", tokenID) // Log for debugging
	return c.token.AccessToken, nil
}

// validateToken simulates token validation
func (c *MockClient) validateToken(token string) error {
	if c.token == nil || c.token.AccessToken != token || c.token.ExpiresAt.Before(time.Now()) {
		return errors.New("invalid or expired token")
	}
	return nil
}

// GetBalance fetches an account balance, requiring a valid token
func (c *MockClient) GetBalance(accountID, token string) (float64, error) {
	if err := c.validateToken(token); err != nil {
		return 0, err
	}
	account, exists := c.accounts[accountID]
	if !exists {
		return 0, fmt.Errorf("Bank Z account %s not found", accountID)
	}
	return account.Balance, nil
}

// InitiateTransfer simulates a transfer to a Bank Z account, requiring a valid token
func (c *MockClient) InitiateTransfer(fromAccountID, toAccountID, token string, amount float64) (string, error) {
	if err := c.validateToken(token); err != nil {
		return "", err
	}
	if amount <= 0 {
		return "", errors.New("amount must be positive")
	}
	// Only validate toAccountID as a Bank Z account
	account, exists := c.accounts[toAccountID]
	if !exists {
		return "", fmt.Errorf("destination Bank Z account %s not found", toAccountID)
	}
	// Update Bank Z account balance for simulation
	c.accounts[toAccountID] = MockAccount{
		AccountID: toAccountID,
		Balance:   account.Balance + amount,
	}
	transactionID := fmt.Sprintf("bankz-tx-%d", time.Now().UnixNano())
	fmt.Printf("Bank Z transfer: from %s to %s, amount %.2f, txID %s\n", fromAccountID, toAccountID, amount, transactionID)
	return transactionID, nil
}
