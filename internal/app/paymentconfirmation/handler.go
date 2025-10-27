package paymentconfirmation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	"twitch-crypto-donations/internal/pkg/middleware"

	"github.com/AlekSi/pointer"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type RpcClient interface {
	GetTransaction(ctx context.Context, txSig solana.Signature, opts *rpc.GetTransactionOpts) (out *rpc.GetTransactionResult, err error)
}

type RequestBody struct {
	Signature string `json:"signature"`
	Recipient string `json:"recipient"`
	SolAmount string `json:"sol_amount"`
}

type ResponseBody struct {
	Confirmed bool   `json:"confirmed"`
	Message   string `json:"message"`
	Slot      uint64 `json:"slot,omitempty"`
}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[ResponseBody]
)

type Handler struct {
	rpcClient RpcClient
}

func New(rpcClient RpcClient) *Handler {
	return &Handler{
		rpcClient: rpcClient,
	}
}

func (h *Handler) Handle(ctx context.Context, request Request) (*Response, error) {
	sig, err := solana.SignatureFromBase58(request.Body.Signature)
	if err != nil {
		return nil, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var tx *rpc.GetTransactionResult

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, errors.New("transaction confirmation timeout after 15 seconds")
		case <-ticker.C:
			tx, err = h.rpcClient.GetTransaction(
				timeoutCtx, sig,
				&rpc.GetTransactionOpts{
					Encoding:                       solana.EncodingBase64,
					Commitment:                     rpc.CommitmentConfirmed,
					MaxSupportedTransactionVersion: pointer.ToUint64(0),
				},
			)

			if err != nil {
				continue
			}

			if tx == nil {
				continue
			}

			if tx.Meta.Err != nil {
				return nil, fmt.Errorf("transaction failed: %+v", tx.Meta.Err)
			}

			return &Response{
				StatusCode: http.StatusOK,
				Body: ResponseBody{
					Confirmed: true,
					Message:   "Transaction successfully confirmed on Solana.",
					Slot:      tx.Slot,
				},
			}, nil
		}
	}
}
