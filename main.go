package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/wrtx-dev/golocalsend/localsend"
	"github.com/wrtx-dev/golocalsend/localsend/config"
	"github.com/wrtx-dev/golocalsend/localsend/proto"
)

const (
	GOLOCALSEND_PATH  = "GOLOCALSEND_PATH"
	GOLOCALSEND_HTTPS = "GOLOCALSEND_HTTPS"
	GOLOCALSEND_PORT  = "GOLOCALSEND_PORT"
	GOLOCALSEND_ALIAS = "GOLOCALSEND_ALIAS"
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
				Value: true,
			},
			&cli.StringFlag{
				Name:  "alias",
				Usage: "set instance's alias",
				Value: "go local send server",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "set listen port,both tcp and udp",
				Value: 53317,
			},
		},
		Action: func(c *cli.Context) error {
			cf, err := config.DefaultLocalsendConfig()
			if err != nil {
				return err
			}
			cf.SavePath = os.Getenv(GOLOCALSEND_PATH)
			if cf.SavePath == "" {
				sp := c.String("path")
				if sp != "" {
					cf.SavePath = sp
				}
			}
			https := false
			httpsFlag := os.Getenv(GOLOCALSEND_HTTPS)
			if httpsFlag == "" {
				https = c.Bool("https")
			} else {
				https = strings.ToLower(httpsFlag) == "true"
			}

			if !https {
				cf.HTTPS = false
				cf.Cert = nil
			}
			alias := os.Getenv(GOLOCALSEND_ALIAS)
			if alias == "" {
				alias = c.String("alias")
			}
			if alias != "" {
				cf.Alias = alias
			}

			portStr := os.Getenv(GOLOCALSEND_PORT)
			if portStr == "" {
				cf.Port = uint(c.Int("port"))
			} else {
				port, err := strconv.Atoi(portStr)
				if err != nil {
					return err
				}
				cf.Port = uint(port)
			}
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
