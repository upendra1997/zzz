package main

import (
	"flag"
	"log/slog"
	"main/api"
	"main/storage"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	storageType := flag.String("storage", "sqlite", "Type of storage to use: 'inmemory' or 'sqlite'")
	sqliteDBFile := flag.String("sqlite_db_file", "", "File path for SQLite database: 'store.db'; defaults to :memory: if empty or invalid path")
	flag.Parse()

	var s storage.Storage
	switch *storageType {
	case "inmemory":
		slog.Info("Using in-memory storage")
		s = storage.NewInMemoryStorage()
	case "sqlite":
		slog.Info("Using SQLite storage", "db_file", *sqliteDBFile)
		s = storage.NewSqliteStorage(*sqliteDBFile)
	default:
		slog.Error("Invalid storage type specified", "storageType", *storageType)
		return
	}

	accountHandler := api.NewAccountHandlers(s)

	router := mux.NewRouter()
	router.HandleFunc("/accounts", accountHandler.CreateAccount).Methods("POST")
	router.HandleFunc("/accounts/{account_id}", accountHandler.GetAccount).Methods("GET")
	router.HandleFunc("/transactions", accountHandler.SubmitTransaction).Methods("POST")
	slog.Info("Starting server on :8080")
	slog.Error("Server Crashed", "error", http.ListenAndServe(":8080", router))
}
