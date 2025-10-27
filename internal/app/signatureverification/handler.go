package signatureverification

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"twitch-crypto-donations/internal/pkg/middleware"

	"github.com/mr-tron/base58"
)

type Database interface {
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

type JwtManager interface {
	GenerateJwt(address string) (string, error)
}

type RequestBody struct {
	Address   string `json:"address"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type ResponseBody struct {
	JwtToken string `json:"jwt_token"`
}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[ResponseBody]
)

type Handler struct {
	db  Database
	jwt JwtManager
}

func New(db Database, jwt JwtManager) *Handler {
	return &Handler{db: db, jwt: jwt}
}

func (h *Handler) Handle(_ context.Context, request Request) (*Response, error) {
	nonce, err := h.extractNonce(request.Body.Message)
	if err != nil {
		return nil, err
	}

	if err = h.validateAndConsumeNonce(nonce, request.Body.Address); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if err = h.verifySolanaSignature(request.Body.Address, request.Body.Signature, request.Body.Message); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	jwtToken, err := h.jwt.GenerateJwt(request.Body.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &Response{
		Body:       ResponseBody{JwtToken: jwtToken},
		StatusCode: http.StatusOK,
	}, nil
}

func (h *Handler) verifySolanaSignature(publicKeyStr, signatureStr, message string) error {
	publicKeyBytes, err := base58.Decode(publicKeyStr)
	if err != nil {
		return fmt.Errorf("invalid public key format: %w", err)
	}

	if len(publicKeyBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key length: expected %d, got %d", ed25519.PublicKeySize, len(publicKeyBytes))
	}

	signatureBytes, err := h.decodeSignature(signatureStr)
	if err != nil {
		return fmt.Errorf("invalid signature format: %w", err)
	}

	if len(signatureBytes) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature length: expected %d, got %d", ed25519.SignatureSize, len(signatureBytes))
	}

	messageBytes := []byte(message)

	valid := ed25519.Verify(publicKeyBytes, messageBytes, signatureBytes)
	if !valid {
		return errors.New("signature verification failed: invalid signature")
	}

	return nil
}

func (h *Handler) decodeSignature(signatureStr string) ([]byte, error) {
	signatureBytes, err := hex.DecodeString(signatureStr)
	if err == nil && len(signatureBytes) == ed25519.SignatureSize {
		return signatureBytes, nil
	}

	return nil, errors.New("signature must be either hex encoded")
}

func (h *Handler) validateAndConsumeNonce(nonce, claimedAddress string) error {
	deleteQuery := `DELETE FROM nonces WHERE nonce = $1 RETURNING address;`

	var storedAddress string
	err := h.db.QueryRow(deleteQuery, nonce).Scan(&storedAddress)

	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("nonce is invalid or has expired")
	}

	if err != nil {
		return fmt.Errorf("database query error: %w", err)
	}

	if !strings.EqualFold(storedAddress, claimedAddress) {
		return errors.New("nonce was requested by a different address")
	}

	return nil
}

func (h *Handler) extractNonce(message string) (string, error) {
	const prefix = "Nonce: "

	start := strings.Index(message, prefix)
	if start == -1 {
		return "", errors.New("nonce not found in message")
	}

	start += len(prefix)

	end := start
	for end < len(message) {
		end++
	}

	if end == start {
		return "", errors.New("empty nonce")
	}

	nonce := strings.TrimSpace(message[start:end])
	if nonce == "" {
		return "", errors.New("empty nonce after trimming")
	}

	return nonce, nil
}
