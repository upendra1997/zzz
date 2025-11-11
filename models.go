package main

type AccountRequest struct {
	AccountId      uint64 `json:"account_id"`
	InitialBalance string `json:"initial_balance"`
}

type AccountResponse struct {
	AccountId uint64 `json:"account_id"`
	Balance   string `json:"balance"`
}

type TransactionRequest struct {
	SourceAccountId      uint64 `json:"source_account_id"`
	DestinationAccountId uint64 `json:"destination_account_id"`
	Amount               string `json:"amount"`
}
