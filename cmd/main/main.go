package main

import (
	"context"
	"os/signal"
	"syscall"

	"medodsTest/internal/app"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	defer cancel()

	app.Start(ctx)

}
