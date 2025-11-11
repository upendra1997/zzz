package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func accounts(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var req AccountRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Account ID: %d, Initial Balance: %s\n", req.AccountId, req.InitialBalance)
	rw.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/accounts", accounts)
	http.ListenAndServe(":8000", nil)
	fmt.Println("Hello World")
}
