package main

import (
	"context"
)

func main() {
	ctx := context.Background()

	server, err := InitializeServer(ctx)
	if err != nil {
		panic(err)
	}

	server.ServerHTTP()
}
