package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"main/api"
	"main/model"
	"main/storage"
	"math"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	// Added for potential delays
)

type FlakyMemory struct {
	*storage.InMemoryStorage
}

type FlakyMemoryTransaction struct {
	*storage.InMemoryStorageTransaction
}

func NewMockInMemory() *FlakyMemory {
	return &FlakyMemory{
		storage.NewInMemoryStorage(),
	}
}

func (ms *FlakyMemory) Begin() storage.StorageTransaction {
	tx := ms.InMemoryStorage.Begin()
	return &FlakyMemoryTransaction{tx.(*storage.InMemoryStorageTransaction)}
}

// Set sets the balance for a given account ID.
func (mt *FlakyMemoryTransaction) Set(accountID uint64, balance string) error {
	if rand.Float64() < 0.01 {
		// Simulate a failure 1% of the time
		return fmt.Errorf("simulated storage failure for account %d", accountID)
	}
	return mt.InMemoryStorageTransaction.Set(accountID, balance)
}

func TestSubmitTransaction_InconsistententBalance_InMemory(t *testing.T) {
	mockStorage := NewMockInMemory()
	handlers := api.NewAccountHandlers(mockStorage)

	// Create initial accounts
	account1ID := uint64(1001)
	account2ID := uint64(1002)
	initialBalance := "1000.000000000" // Use high precision string

	tx := mockStorage.Begin()
	tx.Set(account1ID, initialBalance)
	tx.Set(account2ID, initialBalance)
	tx.Commit()

	numConcurrentTransactions := 1000
	transferAmountStr := "1.000000000" // Each transaction transfers this amount
	// transferAmount, _ := strconv.ParseFloat(transferAmountStr, 64)

	var wg sync.WaitGroup
	wg.Add(numConcurrentTransactions)

	t.Logf("Running %d concurrent transactions from account %d to %d, each transferring %s",
		numConcurrentTransactions, account1ID, account2ID, transferAmountStr)

	for i := range numConcurrentTransactions {
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
				t.Logf("Transaction %d failed with status %d: %s", transactionNum, rr.Code, rr.Body.String())
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

	t.Logf("Initial balance (Account 1): %s", initialBalance)
	t.Logf("Initial balance (Account 2): %s", initialBalance)
	t.Logf("final balance (Account 1): %.9f", finalBalance1Float)
	t.Logf("final balance (Account 2): %.9f", finalBalance2Float)

	epsilon := 1e-3
	initialTotalBalnace := 2000.0
	finalTotalBalance := finalBalance1Float + finalBalance2Float

	if finalBalance1Float >= finalBalance2Float {
		t.Errorf("Account 1 balance should be less than Account 2 balance after transactions")
	}

	if math.Abs(finalTotalBalance-initialTotalBalnace) > epsilon {
		t.Errorf("Total balance mismatch: Expected %f, Got %f",
			initialTotalBalnace, finalTotalBalance)
	}
}
