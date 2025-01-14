package localsend

import (
	"context"

	"github.com/wrtx-dev/golocalsend/localsend/config"
	"github.com/wrtx-dev/golocalsend/localsend/serve"
	"golang.org/x/exp/rand"
)

type LocalsendApp struct {
	Config config.LocalsendConfig
}

func NewApp(config config.LocalsendConfig) (*LocalsendApp, error) {
	return &LocalsendApp{
		Config: config,
	}, nil
}

func (app *LocalsendApp) Run(ctx context.Context, comfirm serve.FComfirmUpload) error {
	server := serve.NewServer(ctx, &app.Config, comfirm)
	go server.Serve()
	return nil
}

const (
	Letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = Letters[rand.Int63()%int64(len(Letters))]
	}
	return string(b)
}
