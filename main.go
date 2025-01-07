package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/wrtx-dev/golocalsend/localsend"
	"github.com/wrtx-dev/golocalsend/localsend/config"
	"github.com/wrtx-dev/golocalsend/localsend/proto"
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
	app, err := localsend.NewApp(*cf)
	if err != nil {
		os.Exit(-1)
	}
	if err = app.Run(ctx, func(files map[string]proto.FileInfo) (map[string]proto.FileInfo, error) {
		res := map[string]proto.FileInfo{}
		for k, v := range files {
			savePath := filepath.Join(cf.SavePath, v.FileName)
			if _, err := os.Stat(savePath); err != nil {
				if os.IsNotExist(err) {
					res[k] = v
				} else {
					fmt.Println("file:", savePath, "exists:", err)
				}
			}
		}
		return res, nil
	}); err != nil {
		os.Exit(-1)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	s := <-c
	fmt.Println("get signal:", s)
	cancel()
}
