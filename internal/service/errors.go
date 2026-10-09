package service

import "errors"

var (
	// ErrEmailTaken indicates registration with an already-used email (409).
	ErrEmailTaken = errors.New("user with this email already exists")
	// ErrInvalidCredentials indicates a failed login (401).
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInsufficientFunds indicates the wallet balance is too low (400).
	ErrInsufficientFunds = errors.New("insufficient funds")
	// ErrInvalidAmount indicates a non-positive amount (400).
	ErrInvalidAmount = errors.New("amount must be greater than zero")
	// ErrRecipientNotFound indicates a transfer target that does not exist (400).
	ErrRecipientNotFound = errors.New("recipient not found")
	// ErrSelfTransfer indicates a transfer to the sender's own account (400).
	ErrSelfTransfer = errors.New("cannot transfer to your own account")
	// ErrWalletNotFound indicates a missing wallet.
	ErrWalletNotFound = errors.New("wallet not found")
)
