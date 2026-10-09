package model

import "time"

// ---- User DTOs ----

type UserRegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserRegisterResponse struct {
	UserID  string `json:"userID"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResponse struct {
	Token string `json:"token"`
}

// ---- Wallet DTOs ----

type WalletBalanceResponse struct {
	UserID  string  `json:"userID"`
	Balance float64 `json:"balance"`
}

// ---- Transaction DTOs ----

type TopUpRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type TopUpResponse struct {
	TransactionID string  `json:"transactionID"`
	Message       string  `json:"message"`
	NewBalance    float64 `json:"newBalance"`
}

type TransferRequest struct {
	ToUserID string  `json:"toUserID" binding:"required,uuid"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
}

type TransferResponse struct {
	TransactionID string `json:"transactionID"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

// ---- Internal (service-to-service) DTOs ----

// WalletDebitRequest is sent by the Payment Service to debit a user's wallet
// for a completed payment. The payment session is owned by the Payment
// Service; core-service only moves money and records the ledger entry.
type WalletDebitRequest struct {
	UserID        string  `json:"userID" binding:"required,uuid"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	MerchantID    string  `json:"merchantID" binding:"required"`
	Description   string  `json:"description"`
	TransactionID string  `json:"transactionID"`
}

// WalletDebitResponse returns the resulting balance and recorded ledger ID.
type WalletDebitResponse struct {
	TransactionID string  `json:"transactionID"`
	NewBalance    float64 `json:"newBalance"`
}

// ---- Transaction history DTOs ----

type TransactionHistoryItem struct {
	TransactionID string    `json:"transactionID"`
	Type          string    `json:"type"`
	Amount        float64   `json:"amount"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
}

type Pagination struct {
	CurrentPage int   `json:"currentPage"`
	TotalPages  int   `json:"totalPages"`
	TotalItems  int64 `json:"totalItems"`
}

type TransactionHistoryResponse struct {
	Transactions []TransactionHistoryItem `json:"transactions"`
	Pagination   Pagination               `json:"pagination"`
}

// ---- Common error DTO ----

type ErrorResponse struct {
	Error string `json:"error"`
}
