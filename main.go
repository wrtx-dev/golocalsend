package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrtx-dev/golocalsend/localsend/serve"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go serve.ServeUDP(ctx)
	go serve.RegisterHTTPClient(ctx)
	go serve.HandleFileServer(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	s := <-c
	fmt.Println("get signal:", s)
	cancel()
}
