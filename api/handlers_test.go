package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"main/model"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time" // Added for potential delays
)

// MockStorage implements the storage.Storage interface for testing purposes.
// It deliberately does NOT use any locks to expose race conditions.
type MockStorage struct {
	accounts map[uint64]string
}

// NewMockStorage creates a new MockStorage instance.
func NewMockStorage() *MockStorage {
	return &MockStorage{
		accounts: make(map[uint64]string),
	}
}

// Get retrieves the balance for a given account ID.
func (ms *MockStorage) Get(accountID uint64) (string, error) {
	balance, ok := ms.accounts[accountID]
	if !ok {
		return "", fmt.Errorf("account %d not found", accountID)
	}
	return balance, nil
}

// Set sets the balance for a given account ID.
func (ms *MockStorage) Set(accountID uint64, balance string) error {
	// Simulate a small delay to increase the chance of race conditions
	time.Sleep(1 * time.Millisecond)
	ms.accounts[accountID] = balance
	return nil
}

func (ms *MockStorage) Delete(accountID uint64) error {
	return nil
}

// TestSubmitTransaction_RaceCondition tests for race conditions in SubmitTransaction.
// This test is designed to be run with the Go race detector: `go test -race ./...`
func TestSubmitTransaction_RaceCondition(t *testing.T) {
	// Initialize mock storage without locks
	mockStorage := NewMockStorage()
	handlers := NewAccountHandlers(mockStorage)

	// Create initial accounts
	account1ID := uint64(1001)
	account2ID := uint64(1002)
	initialBalance := "1000.000000000" // Use high precision string

	mockStorage.Set(account1ID, initialBalance)
	mockStorage.Set(account2ID, initialBalance)

	numConcurrentTransactions := 1000
	transferAmountStr := "1.000000000" // Each transaction transfers this amount
	// transferAmount, _ := strconv.ParseFloat(transferAmountStr, 64)

	var wg sync.WaitGroup
	wg.Add(numConcurrentTransactions)

	t.Logf("Running %d concurrent transactions from account %d to %d, each transferring %s",
		numConcurrentTransactions, account1ID, account2ID, transferAmountStr)

	for i := 0; i < numConcurrentTransactions; i++ {
		go func(transactionNum int) {
			defer wg.Done()

			reqBody := model.TransactionRequest{
				SourceAccountId:      account1ID,
				DestinationAccountId: account2ID,
				Amount:               transferAmountStr,
			}
			bodyBytes, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handlers.SubmitTransaction(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Transaction %d failed with status %d: %s", transactionNum, rr.Code, rr.Body.String())
			}
		}(i)
	}

	wg.Wait()

	// Verify final balances after all transactions
	finalBalance1Str, err := mockStorage.Get(account1ID)
	if err != nil {
		t.Fatalf("Failed to get final balance for account %d: %v", account1ID, err)
	}
	finalBalance2Str, err := mockStorage.Get(account2ID)
	if err != nil {
		t.Fatalf("Failed to get final balance for account %d: %v", account2ID, err)
	}

	finalBalance1Float, _ := strconv.ParseFloat(finalBalance1Str, 64)
	finalBalance2Float, _ := strconv.ParseFloat(finalBalance2Str, 64)

	expectedFinalBalance1 := 0.0
	expectedFinalBalance2 := 2000.0

	t.Logf("Initial balance (Account 1): %s", initialBalance)
	t.Logf("Initial balance (Account 2): %s", initialBalance)
	t.Logf("Expected final balance (Account 1): %.9f", expectedFinalBalance1)
	t.Logf("Actual final balance (Account 1): %.9f", finalBalance1Float)
	t.Logf("Expected final balance (Account 2): %.9f", expectedFinalBalance2)
	t.Logf("Actual final balance (Account 2): %.9f", finalBalance2Float)

	epsilon := 1e-3

	if math.Abs(finalBalance1Float-expectedFinalBalance1) > epsilon {
		t.Errorf("Account %d balance mismatch: Expected %f, Got %f",
			account1ID, expectedFinalBalance1, finalBalance1Float)
	}

	if math.Abs(finalBalance2Float-expectedFinalBalance2) > epsilon {
		t.Errorf("Account %d balance mismatch: Expected %f, Got %f",
			account2ID, expectedFinalBalance2, finalBalance2Float)
	}
}
