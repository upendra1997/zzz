# Simple Go Banking API

This project implements a basic RESTful API for managing bank accounts and processing transactions using Go. Utilizing an in-memory data store for simplicity.

## Simple Approach

The core idea is to provide endpoints for common banking operations: creating accounts, checking balances, and transferring funds. Accounts are stored in a basic in-memory map. Transactions involve debiting a source account and crediting a destination account. A key focus is demonstrating how concurrency can lead to issues if not properly managed.

## Features
*   **Account Management**:
    *   Create new bank accounts with an initial balance.
    *   Retrieve the current balance for any given account.
*   **Transaction Processing**:
    *   Submit transactions to transfer funds between two accounts.
    *   Basic validation for transaction amounts and account existence.
*   **In-Memory Storage**: Uses a simple in-memory map for account data, making it easy to set up and test without external database dependencies.

## Setup Instructions

### Prerequisites

Before you begin, ensure you have the following installed:

*   **Go**: Version 1.18 or higher. You can download it from [go.dev](https://go.dev/dl/).

### Getting Started

1.  **Install Dependencies**:
    The project uses Go modules to manage dependencies. From the project root directory, run:
    ```bash
    go mod tidy
    ```
    This command will download any required packages, such as `github.com/gorilla/mux`.

2.  **Run the Application**:
    To start the API server, execute the `main.go` file from the project root:
    ```bash
    go run main.go
    ```
    By default, the API server will likely start on `http://localhost:8080`. You can then use tools like `curl` or Postman to interact with the endpoints.

3.  **Run Tests (including Race Detector)**:
    It's crucial to run tests with the Go race detector enabled to catch concurrency issues. From the project root:
    ```bash
    go test -race ./...
    ```
    This command will run all tests in the project (including those in `api/handlers_test.go` that specifically target race conditions) and report any data races detected. This is a powerful feature for identifying potential bugs in concurrent Go applications.

## API Endpoints

Here are the primary API endpoints provided by this service:

*   **Create Account**
    *   **Method**: `POST`
    *   **Path**: `/accounts`
    *   **Request Body**:
        ```json
        {
            "account_id": 123,
            "initial_balance": "100.23"
        }
        ```
    *   **Example `curl` command**:
        ```bash
        curl -X POST -H "Content-Type: application/json" \
             -d '{"account_id": 1, "initial_balance": "1000.00"}' \
             http://localhost:8080/accounts
        ```
        ```bash
        curl -X POST -H "Content-Type: application/json" \
             -d '{"account_id": 2, "initial_balance": "500.00"}' \
             http://localhost:8080/accounts
        ```
    *   **Response**: `200 OK` (empty body on success) or an error.

*   **Get Account Details**
    *   **Method**: `GET`
    *   **Path**: `/accounts/{account_id}`
    *   **Example**: `/accounts/123`
    *   **Example `curl` command**:
        ```bash
        curl http://localhost:8080/accounts/1
        ```
    *   **Response**:
        ```json
        {
            "account_id": 123,
            "balance": "100.230000000"
        }
        ```
        or `404 Not Found` if the account does not exist.

*   **Submit Transaction**
    *   **Method**: `POST`
    *   **Path**: `/transactions`
    *   **Request Body**:
        ```json
        {
            "source_account_id": 123,
            "destination_account_id": 456,
            "amount": "10.00"
        }
        ```
    *   **Example `curl` command**:
        ```bash
        curl -X POST -H "Content-Type: application/json" \
             -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "150.00"}' \
             http://localhost:8080/transactions
        ```
    *   **Response**: `200 OK` (empty body on success) or an error (e.g., insufficient funds, account not found).
