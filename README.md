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

### Approaches
#### Race Conditions:
1.  **Simple Approach**: The initial implementation uses a straightforward method to handle account and transaction management. However, this approach does not include concurrency controls, which lead to race conditions when multiple transactions are processed simultaneously.
2. **Gloabal Locking Approach**: An improved version that introduces a global mutex to serialize access to the account data store. This approach prevents race conditions by ensuring that only one transaction can modify the account balances at a time, albeit at the cost of reduced concurrency and potential performance bottlenecks.
```go
type AccountHandlers struct {
	storage storage.Storage
	lock    sync.RWMutex
}
```
3. **Fine-Grained Locking Approach**: A more sophisticated solution that employs per-account mutexes to allow concurrent transactions on different accounts while still preventing race conditions. This approach balances concurrency and data integrity by locking only the accounts involved in a transaction.

#### Atomicity and Consistency:
1.  **Simple Approach**: Transactions are processed without ensuring atomicity, which can lead to inconsistent states if a failure occurs mid-transaction.
2. **Two-Phase Commit Simulation**: An enhanced method that simulates a two-phase commit protocol to ensure that both the debit and credit operations of a transaction are completed successfully or not at all. This approach helps maintain data consistency even in the face of errors or failures during transaction processing. We can do that by adding a rollback in case of failure.
```go
// Attempt to update destination account
err = h.storage.Set(req.DestinationAccountId, fmt.Sprintf(SPRINTF_FORMAT, newDestinationBalance))
if err != nil {
	// Rollback source account if destination update fails
	rollbackErr := h.storage.Set(req.SourceAccountId, sourceBalance)
	if rollbackErr != nil {
		// Log this error, as the system is now in an inconsistent state
		// For simplicity, we'll return a generic error, but in a real system,
		// this would trigger an alert for manual intervention.
		http.Error(rw, fmt.Sprintf("Failed to update destination account balance and rollback failed: %s", rollbackErr.Error()), http.StatusInternalServerError)
		return
	}
}
```
#### concurrency:
1.  **Simple Approach**: Lacks any concurrency controls, leading to potential data races and incorrect balances when multiple transactions are processed simultaneously.
2. **Gloabal Locking Approach**: Introduces a global mutex to serialize access to the account data store, preventing race conditions, but at the cost of reduced concurrency.
3. **Isolation**: Each transaction operate on local copy of the account balances, and save the changes to the main balance once the transaction is successful. This way, concurrent transactions do not interfere with each other until they are ready to commit their changes, and mulitple transactions can be processed in parallel.
