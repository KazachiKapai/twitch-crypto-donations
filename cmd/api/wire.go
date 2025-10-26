//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"twitch-crypto-donations/internal/config"
	"twitch-crypto-donations/internal/pkg/server"

	"github.com/google/wire"
)

func InitializeServer(ctx context.Context) (*server.Server, error) {
	wire.Build(config.WireSet)

	return nil, nil
}
