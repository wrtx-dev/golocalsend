package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/wrtx-dev/golocalsend/localsend"
	"github.com/wrtx-dev/golocalsend/localsend/config"
	"github.com/wrtx-dev/golocalsend/localsend/proto"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	App := cli.App{
		Name:  "golocalsend",
		Usage: "A program that partially implements the LocalSend protocol.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "set save path",
				Value: "./",
			},
			&cli.BoolFlag{
				Name:  "https",
				Usage: "use https request/response",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "alias",
				Usage: "set instance's alias",
				Value: "go local send server",
			},
		},
		Action: func(c *cli.Context) error {
			cf, err := config.DefaultLocalsendConfig()
			if err != nil {
				return err
			}
			sp := c.String("path")
			cf.SavePath = sp

			https := c.Bool("https")
			if !https {
				cf.HTTPS = false
				cf.Cert = nil
			}
			alias := c.String("alias")
			cf.Alias = alias
			app, err := localsend.NewApp(*cf)
			if err != nil {
				return err
			}
			err = app.Run(ctx, func(files map[string]proto.FileInfo) (map[string]proto.FileInfo, error) {
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
			})
			if err != nil {
				return err
			}
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
			<-ch
			cancel()
			return nil
		},
	}
	if err := App.Run(os.Args); err != nil {
		os.Exit(-1)
	}
}
