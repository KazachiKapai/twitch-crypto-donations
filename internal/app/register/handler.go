package register

import (
	"context"
	"fmt"
	"twitch-crypto-donations/internal/pkg/middleware"
)

type Database interface {
}

type RequestBody struct{}

type ResponseBody struct{}

type (
	Request  = middleware.Request[RequestBody]
	Response = middleware.Response[ResponseBody]
)

type Handler struct {
	db Database
}

func New(db Database) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) Handle(ctx context.Context, request Request) (*Response, error) {
	fmt.Print("HELLO")
	return nil, nil
}
