package serve

import (
	"context"
	"net"

	"github.com/wrtx-dev/golocalsend/localsend/config"
	"github.com/wrtx-dev/golocalsend/localsend/proto"
)

type GoLocalsendServer struct {
	ctx    context.Context
	config config.LocalsendConfig
}

type NewClientChan struct {
	request *proto.RegisterRequest
	addr    net.Addr
}

func NewServer(ctx context.Context, config *config.LocalsendConfig) *GoLocalsendServer {
	return &GoLocalsendServer{
		ctx:    ctx,
		config: *config,
	}
}

func (s *GoLocalsendServer) Serve() error {
	protoStr := "https"
	if !s.config.HTTPS {
		protoStr = "http"
	}
	machineInfo := &proto.RegisterRequest{
		Alias:        s.config.Alias,
		Version:      "2.0",
		DeviceModel:  s.config.DeviceModel,
		Port:         int(s.config.Port),
		FingerPrint:  s.config.RandStr,
		Protocol:     protoStr,
		Download:     false,
		Announcement: true,
		Announce:     true,
	}

	go s.fileRecorder()
	go s.ServeUDP()
	go s.RegisterHTTPClient(machineInfo)
	go s.HandleFileServer()
	return nil
}
