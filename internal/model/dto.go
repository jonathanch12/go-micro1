package model

import "time"

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

type WalletBalanceResponse struct {
	UserID  string  `json:"userID"`
	Balance float64 `json:"balance"`
}

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

type PayRequest struct {
	MerchantID  string  `json:"merchantID" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type PayResponse struct {
	TransactionID string  `json:"transactionID"`
	Message       string  `json:"message"`
	NewBalance    float64 `json:"newBalance"`
}

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

type ErrorResponse struct {
	Error string `json:"error"`
}
