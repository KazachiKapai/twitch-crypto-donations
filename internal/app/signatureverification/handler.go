package signatureverification

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"twitch-crypto-donations/internal/pkg/middleware"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

	recoveredAddress, err := h.getRecoveredAddress(request.Body.Signature, request.Body.Message)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(recoveredAddress) != strings.ToLower(request.Body.Address) {
		return nil, fmt.Errorf("signature verification failed: address mismatch. Recovered: %s", recoveredAddress)
	}

	jwtToken, err := h.jwt.GenerateJwt(request.Body.Address)
	if err != nil {
		return nil, err
	}

	return &Response{Body: ResponseBody{JwtToken: jwtToken}, StatusCode: http.StatusOK}, nil
}

func (h *Handler) getRecoveredAddress(signature, message string) (string, error) {
	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	msgHash := crypto.Keccak256Hash([]byte(prefixedMessage))

	sig, err := hexutil.Decode(signature)
	if err != nil {
		return "", err
	}

	if len(sig) != 65 {
		return "", errors.New("invalid signature length")
	}

	if sig[64] == 27 || sig[64] == 28 {
		sig[64] -= 27
	}

	pubKeyRaw, err := crypto.Ecrecover(msgHash.Bytes(), sig)
	if err != nil {
		return "", errors.New("signature recovery failed")
	}

	pubKey, err := crypto.UnmarshalPubkey(pubKeyRaw)
	if err != nil {
		return "", errors.New("invalid public key")
	}

	return crypto.PubkeyToAddress(*pubKey).Hex(), nil
}

func (h *Handler) validateAndConsumeNonce(nonce, claimedAddress string) error {
	var storedAddress string

	selectQuery := "SELECT address FROM nonces WHERE nonce = $1;"
	err := h.db.QueryRow(selectQuery, nonce).Scan(&storedAddress)

	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("nonce is invalid or has expired")
	}

	if err != nil {
		return fmt.Errorf("database query error: %w", err)
	}

	if strings.ToLower(storedAddress) != strings.ToLower(claimedAddress) {
		return errors.New("nonce requested by different address")
	}

	deleteQuery := "DELETE FROM nonces WHERE nonce = $1;"
	result, err := h.db.Exec(deleteQuery, nonce)
	if err != nil {
		return fmt.Errorf("failed to consume nonce: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("failed to consume nonce: no rows affected")
	}

	return nil
}

func (h *Handler) extractNonce(message string) (string, error) {
	start := strings.Index(message, "Nonce: ")
	if start == -1 {
		return "", errors.New("nonce not found in message")
	}

	start += len("Nonce: ")

	end := strings.Index(message[start:], ",")
	if end == -1 {
		return "", errors.New("nonce format incorrect")
	}

	return message[start : start+end], nil
}
