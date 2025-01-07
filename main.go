package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrtx-dev/golocalsend/localsend"
	"github.com/wrtx-dev/golocalsend/localsend/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	// app, err := localsend.NewApp(config.LocalsendConfig{
	// 	HTTPS:         false,
	// 	Port:          53317,
	// 	MulticastAddr: "224.0.0.167",
	// 	Alias:         "go local send test",
	// })
	cf, err := config.DefaultLocalsendConfig()
	if err != nil {
		os.Exit(-1)
	}
	cf.SavePath = "/tmp/"
	app, err := localsend.NewApp(*cf)
	if err != nil {
		os.Exit(-1)
	}
	if err = app.Run(ctx); err != nil {
		os.Exit(-1)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	s := <-c
	fmt.Println("get signal:", s)
	cancel()
}
