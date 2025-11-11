package api

import (
	"encoding/json"
	"fmt"
	"io" // Added for transaction logging
	"main/model"
	"main/storage"
	"math/big"
	"sync"

	"net/http"
	"strconv"

	"github.com/gorilla/mux" // Using mux for more advanced routing, especially for path variables
)

const PRECISION uint = 64

var SPRINTF_FORMAT = "%.19f"

// AccountHandlers provides HTTP handlers for account-related operations.
type AccountHandlers struct {
	storage storage.Storage
	lock    sync.RWMutex
}

// NewAccountHandlers creates and returns a new AccountHandlers instance.
func NewAccountHandlers(s storage.Storage) *AccountHandlers {
	return &AccountHandlers{storage: s}
}

// CreateAccount handles POST requests to create a new account.
// Request Body: {"account_id": 123, "initial_balance": "100.23"}
// Response: Empty or error
func (h *AccountHandlers) CreateAccount(rw http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var req model.AccountRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(rw, "Invalid request body format", http.StatusBadRequest)
		return
	}
	initialBalanceFloat, _, err := big.ParseFloat(req.InitialBalance, 10, PRECISION, big.ToNearestEven)
	if initialBalanceFloat.Cmp(big.NewFloat(0.0)) < 0 || err != nil {
		http.Error(rw, "Invalid Initial Balance", http.StatusBadRequest)
		return
	}

	initialBalanceStr := fmt.Sprintf(SPRINTF_FORMAT, initialBalanceFloat)

	h.lock.Lock()
	defer h.lock.Unlock()
	tx := h.storage.Begin()
	defer tx.Rollback()
	err = tx.Set(req.AccountId, initialBalanceStr)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusConflict) // Using StatusConflict for existing account
		return
	}
	tx.Commit()

	rw.WriteHeader(http.StatusOK)
}

// GetAccount handles GET requests to retrieve account details.
// Response: {"account_id": 123, "balance": "100.23"} or error
func (h *AccountHandlers) GetAccount(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountIDStr := vars["account_id"]
	accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
	if err != nil {
		http.Error(rw, "Invalid account ID", http.StatusBadRequest)
		return
	}

	h.lock.RLock()
	defer h.lock.RUnlock()
	balance, err := h.storage.Get(accountID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}

	resp := model.AccountResponse{
		AccountId: accountID,
		Balance:   balance,
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(resp)
}

// SubmitTransaction handles POST requests to process transactions.
// Request Body: {"source_account_id": 123, "destination_account_id": 456, "amount": "100.12"}
// Response: Empty or error
func (h *AccountHandlers) SubmitTransaction(rw http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var req model.TransactionRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(rw, "Invalid request body format", http.StatusBadRequest)
		return
	}

	amountFloat, _, err := big.ParseFloat(req.Amount, 10, PRECISION, big.ToNearestEven)
	if err != nil || amountFloat.Cmp(big.NewFloat(0.0)) <= 0 {
		http.Error(rw, "Invalid transaction amount", http.StatusBadRequest)
		return
	}

	if req.SourceAccountId == req.DestinationAccountId {
		http.Error(rw, "Source and destination accounts cannot be the same", http.StatusBadRequest)
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()
	sourceBalance, err := h.storage.Get(req.SourceAccountId)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Source account not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	destinationBalance, err := h.storage.Get(req.DestinationAccountId)
	if err != nil {
		http.Error(rw, fmt.Sprintf("Destination account not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	sourceBalanceAmount, _, err := big.ParseFloat(sourceBalance, 10, PRECISION, big.ToNearestEven)
	destinationBalanceAmount, _, err := big.ParseFloat(destinationBalance, 10, PRECISION, big.ToNearestEven)

	if sourceBalanceAmount.Cmp(amountFloat) < 0 {
		err = fmt.Errorf("insufficient funds in source account")
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	newSourceBalance := sourceBalanceAmount.Sub(sourceBalanceAmount, amountFloat)
	newDestinationBalance := destinationBalanceAmount.Add(destinationBalanceAmount, amountFloat)

	tx := h.storage.Begin()
	defer tx.Rollback()
	err = tx.Set(req.SourceAccountId, fmt.Sprintf(SPRINTF_FORMAT, newSourceBalance))
	if err != nil {
		http.Error(rw, fmt.Sprintf("Failed to update source account balance: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	err = tx.Set(req.DestinationAccountId, fmt.Sprintf(SPRINTF_FORMAT, newDestinationBalance))
	if err != nil {
		http.Error(rw, fmt.Sprintf("Failed to update destination account balance: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	tx.Commit()

	rw.WriteHeader(http.StatusOK)
}
