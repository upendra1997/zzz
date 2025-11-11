package main

import (
	"log/slog"
	"main/api"
	"main/storage"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// storage := storage.NewInMemoryStorage()
	storage := storage.NewSqliteStorage()
	accountHandler := api.NewAccountHandlers(storage)

	router := mux.NewRouter()
	router.HandleFunc("/accounts", accountHandler.CreateAccount).Methods("POST")
	router.HandleFunc("/accounts/{account_id}", accountHandler.GetAccount).Methods("GET")
	router.HandleFunc("/transactions", accountHandler.SubmitTransaction).Methods("POST")
	slog.Info("Starting server on :8080")
	slog.Error("Server Crashed", "error", http.ListenAndServe(":8080", router))
}
