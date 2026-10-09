package coreclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"payment-service/internal/model"
)

// Common errors mapped from core-service responses.
var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrBadRequest        = errors.New("bad request")
	ErrUnauthorized      = errors.New("internal authorization failed")
)

// Client calls the core-service internal API to debit wallets.
type Client struct {
	baseURL     string
	internalKey string
	http        *http.Client
}

// New constructs a core-service client.
func New(baseURL, internalKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL:     baseURL,
		internalKey: internalKey,
		http:        &http.Client{Timeout: timeout},
	}
}

// DebitWallet calls POST /internal/wallets/debit on core-service. It returns
// the ledger transaction ID and the user's new balance.
func (c *Client) DebitWallet(req model.WalletDebitRequest) (*model.WalletDebitResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/internal/wallets/debit",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Internal-Token", c.internalKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call core-service: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var out model.WalletDebitResponse
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("decode debit response: %w", err)
		}
		return &out, nil
	case http.StatusBadRequest:
		if containsInsufficient(data) {
			return nil, ErrInsufficientFunds
		}
		return nil, ErrBadRequest
	case http.StatusNotFound:
		return nil, ErrWalletNotFound
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	default:
		return nil, fmt.Errorf("core-service returned status %d: %s", resp.StatusCode, string(data))
	}
}

func containsInsufficient(body []byte) bool {
	var e model.ErrorResponse
	if err := json.Unmarshal(body, &e); err != nil {
		return false
	}
	return e.Error == ErrInsufficientFunds.Error()
}
