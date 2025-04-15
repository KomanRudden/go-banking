package main

import (
	"fmt"
	"go-banking/bankz"
	"go-banking/models"
	"go-banking/store"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TransferRequest struct {
	FromAccountID string  `json:"fromAccountId"`
	ToAccountID   string  `json:"toAccountId"`
	Amount        float64 `json:"amount"`
}

type BankZBalanceResponse struct {
	AccountID string  `json:"accountId"`
	Balance   float64 `json:"balance"`
}

type TransactionResponse struct {
	ID            string  `json:"id"`
	AccountID     string  `json:"accountId"`
	CustomerID    string  `json:"customerId"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	FromAccountID string  `json:"fromAccountId,omitempty"`
	ToAccountID   string  `json:"toAccountId,omitempty"`
	CreatedAt     string  `json:"createdAt"`
}

var bankzClient = bankz.NewMockClient() // Global client for simplicity

func main() {

	// Gin is a web framework that is designed for building high-performance RESTful APIs.
	// It provides features such as routing, middleware support,
	// JSON validation, and more, making it a popular choice for web development in Go.
	// The `gin.Default()` function initializes a new Gin router instance with default middleware
	// (Logger and Recovery).
	r := gin.Default()

	// Define routes
	r.POST("/api/customers", createCustomer)
	r.GET("/api/customers/:customerId/accounts", getCustomerAccounts)
	r.POST("/api/customers/:customerId/transfers", transferMoney)
	r.GET("/api/customers/:customerId/bankz/balances", getBankZBalances)
	r.GET("/api/customers/:customerId/transactions", getCustomerTransactions)

	// Start the server
	fmt.Println("Starting go-banking server on :8080...")
	if err := r.Run(":8080"); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func createCustomer(c *gin.Context) {
	var req CustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []string{"Invalid request body"}})
		return
	}

	var errors []string
	if req.Name == "" {
		errors = append(errors, "Name cannot be empty")
	}
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s]+$`)
	if req.Name != "" && !nameRegex.MatchString(req.Name) {
		errors = append(errors, "Name must contain only letters and spaces")
	}
	if req.Email == "" {
		errors = append(errors, "Email cannot be empty")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if req.Email != "" && !emailRegex.MatchString(req.Email) {
		errors = append(errors, "Invalid email format")
	}
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	customerID := uuid.New().String()
	currentAccountID := uuid.New().String()
	savingsAccountID := uuid.New().String()
	bonusTransactionID := uuid.New().String()
	now := time.Now()

	customer := models.Customer{
		ID:    customerID,
		Name:  req.Name,
		Email: req.Email,
	}
	store.GlobalStore.AddCustomer(customer)

	currentAccount := models.Account{
		ID:         currentAccountID,
		CustomerID: customerID,
		Type:       "current",
		Balance:    0.0,
		CreatedAt:  now,
	}
	store.GlobalStore.AddAccount(currentAccount)

	savingsAccount := models.Account{
		ID:         savingsAccountID,
		CustomerID: customerID,
		Type:       "savings",
		Balance:    500.0,
		CreatedAt:  now,
	}
	store.GlobalStore.AddAccount(savingsAccount)

	bonusTransaction := models.Transaction{
		ID:         bonusTransactionID,
		AccountID:  savingsAccountID,
		CustomerID: customerID,
		Type:       "bonus",
		Amount:     500.0,
		CreatedAt:  now,
	}
	store.GlobalStore.AddTransaction(bonusTransaction)

	resp := struct {
		CustomerID       string `json:"customerId"`
		CurrentAccountID string `json:"currentAccountId"`
		SavingsAccountID string `json:"savingsAccountId"`
	}{
		CustomerID:       customerID,
		CurrentAccountID: currentAccountID,
		SavingsAccountID: savingsAccountID,
	}
	fmt.Printf("Created customer: %+v\n", resp)
	c.JSON(http.StatusCreated, resp)
}

func getCustomerAccounts(c *gin.Context) {
	customerID := c.Param("customerId")

	if _, found := store.GlobalStore.GetCustomerByID(customerID); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	accounts := store.GlobalStore.GetAccountsByCustomerID(customerID)
	if len(accounts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No accounts found for customer"})
		return
	}

	resp := struct {
		CustomerID string           `json:"customerId"`
		Accounts   []models.Account `json:"accounts"`
	}{
		CustomerID: customerID,
		Accounts:   accounts,
	}
	fmt.Printf("Fetched accounts for customer %s: %+v\n", customerID, accounts)
	c.JSON(http.StatusOK, resp)
}

func transferMoney(c *gin.Context) {
	customerID := c.Param("customerId")
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": []string{"Invalid request body"}})
		return
	}

	var errors []string
	if req.FromAccountID == "" {
		errors = append(errors, "From account ID cannot be empty")
	}
	if req.ToAccountID == "" {
		errors = append(errors, "To account ID cannot be empty")
	}
	if req.Amount <= 0 {
		errors = append(errors, "Amount must be positive")
	}
	if req.FromAccountID == req.ToAccountID && req.FromAccountID != "" {
		errors = append(errors, "Cannot transfer to the same account")
	}
	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if _, found := store.GlobalStore.GetCustomerByID(customerID); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Get OAuth token
	token, err := bankzClient.GetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with Bank Z"})
		return
	}

	// Check fromAccount
	fromAccount, fromFound := store.GlobalStore.GetAccountByID(req.FromAccountID)
	fmt.Printf("Looking up fromAccountID %s: found=%v\n", req.FromAccountID, fromFound)
	if !fromFound {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Source account %s not found", req.FromAccountID)})
		return
	}
	if fromAccount.CustomerID != customerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source account does not belong to the customer"})
		return
	}

	// Check toAccount
	isExternalTransfer := false
	toAccount, toFound := store.GlobalStore.GetAccountByID(req.ToAccountID)
	fmt.Printf("Looking up toAccountID %s: found=%v\n", req.ToAccountID, toFound)
	if !toFound {
		// Try Bank Z account
		_, err := bankzClient.GetBalance(req.ToAccountID, token)
		if err == nil {
			isExternalTransfer = true
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Destination account %s not found", req.ToAccountID)})
			return
		}
	} else if toAccount.CustomerID != customerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Destination account does not belong to the customer"})
		return
	}

	if isExternalTransfer {
		// Bank Z transfer
		if fromAccount.Balance < req.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
			return
		}
		transactionID, err := bankzClient.InitiateTransfer(fromAccount.ID, req.ToAccountID, token, req.Amount)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Update local account balance
		fromAccount.Balance -= req.Amount
		store.GlobalStore.UpdateAccount(fromAccount)

		// Record transaction
		now := time.Now()
		store.GlobalStore.AddTransaction(models.Transaction{
			ID:            transactionID,
			AccountID:     fromAccount.ID,
			CustomerID:    customerID,
			Type:          "bankz_transfer",
			Amount:        req.Amount,
			FromAccountID: fromAccount.ID,
			ToAccountID:   req.ToAccountID,
			CreatedAt:     now,
		})

		resp := struct {
			TransactionID string `json:"transactionId"`
			Status        string `json:"status"`
		}{
			TransactionID: transactionID,
			Status:        "success",
		}
		fmt.Printf("Bank Z transfer successful: %+v\n", resp)
		c.JSON(http.StatusCreated, resp)
		return
	}

	// Internal transfer
	fee := 0.0
	if fromAccount.Type == "current" {
		fee = req.Amount * 0.0005
	}

	totalDeduction := req.Amount + fee
	if fromAccount.Balance < totalDeduction {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
		return
	}

	interest := 0.0
	if toAccount.Type == "savings" {
		interest = req.Amount * 0.005
	}

	fromAccount.Balance -= totalDeduction
	toAccount.Balance += req.Amount + interest
	store.GlobalStore.UpdateAccount(fromAccount)
	store.GlobalStore.UpdateAccount(toAccount)

	now := time.Now()
	transferTransactionID := uuid.New().String()
	store.GlobalStore.AddTransaction(models.Transaction{
		ID:            transferTransactionID,
		AccountID:     fromAccount.ID,
		CustomerID:    customerID,
		Type:          "transfer",
		Amount:        req.Amount,
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		CreatedAt:     now,
	})

	if fee > 0 {
		feeTransactionID := uuid.New().String()
		store.GlobalStore.AddTransaction(models.Transaction{
			ID:         feeTransactionID,
			AccountID:  fromAccount.ID,
			CustomerID: customerID,
			Type:       "fee",
			Amount:     fee,
			CreatedAt:  now,
		})
	}

	if interest > 0 {
		interestTransactionID := uuid.New().String()
		store.GlobalStore.AddTransaction(models.Transaction{
			ID:         interestTransactionID,
			AccountID:  toAccount.ID,
			CustomerID: customerID,
			Type:       "interest",
			Amount:     interest,
			CreatedAt:  now,
		})
	}

	resp := struct {
		TransactionID string `json:"transactionId"`
		Status        string `json:"status"`
	}{
		TransactionID: transferTransactionID,
		Status:        "success",
	}
	fmt.Printf("Internal transfer successful: %+v\n", resp)
	c.JSON(http.StatusCreated, resp)
}

func getBankZBalances(c *gin.Context) {
	customerID := c.Param("customerId")

	if _, found := store.GlobalStore.GetCustomerByID(customerID); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Get OAuth token
	token, err := bankzClient.GetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with Bank Z"})
		return
	}

	linkedAccountIDs := []string{"bankz-acc-123", "bankz-acc-456"}

	var balances []BankZBalanceResponse
	for _, accountID := range linkedAccountIDs {
		balance, err := bankzClient.GetBalance(accountID, token)
		if err != nil {
			continue
		}
		balances = append(balances, BankZBalanceResponse{
			AccountID: accountID,
			Balance:   balance,
		})
	}

	resp := struct {
		CustomerID string                 `json:"customerId"`
		Balances   []BankZBalanceResponse `json:"balances"`
	}{
		CustomerID: customerID,
		Balances:   balances,
	}
	c.JSON(http.StatusOK, resp)
}

func getCustomerTransactions(c *gin.Context) {
	customerID := c.Param("customerId")

	// Verify customer exists
	if _, found := store.GlobalStore.GetCustomerByID(customerID); !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Get query parameters for filtering
	accountID := c.Query("accountId")
	transactionType := c.Query("type")

	// Fetch transactions
	transactions := store.GlobalStore.GetTransactionsByCustomerID(customerID)
	if len(transactions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"customerId":   customerID,
			"transactions": []TransactionResponse{},
		})
		return
	}

	// Filter transactions
	var filtered []models.Transaction
	for _, tx := range transactions {
		if accountID != "" && tx.AccountID != accountID {
			continue
		}
		if transactionType != "" && tx.Type != transactionType {
			continue
		}
		filtered = append(filtered, tx)
	}

	// Convert to response format
	responseTransactions := make([]TransactionResponse, len(filtered))
	for i, tx := range filtered {
		responseTransactions[i] = TransactionResponse{
			ID:            tx.ID,
			AccountID:     tx.AccountID,
			CustomerID:    tx.CustomerID,
			Type:          tx.Type,
			Amount:        tx.Amount,
			FromAccountID: tx.FromAccountID,
			ToAccountID:   tx.ToAccountID,
			CreatedAt:     tx.CreatedAt.Format(time.RFC3339),
		}
	}

	resp := struct {
		CustomerID   string                `json:"customerId"`
		Transactions []TransactionResponse `json:"transactions"`
	}{
		CustomerID:   customerID,
		Transactions: responseTransactions,
	}
	fmt.Printf("Fetched %d transactions for customer %s\n", len(responseTransactions), customerID)
	c.JSON(http.StatusOK, resp)
}
