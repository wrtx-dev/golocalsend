package serve

import (
	"fmt"
	"net"
	"strings"

	"github.com/wrtx-dev/golocalsend/localsend/client"
	"github.com/wrtx-dev/golocalsend/localsend/proto"
)

func (s *GoLocalsendServer) RegisterHTTPClient(machineInfo *proto.RegisterRequest) {
	for {
		select {
		case req := <-clientChan:
			err := client.RegisterClient(machineInfo, req.addr.(*net.UDPAddr).IP.String(), uint64(req.request.Port), strings.ToLower(req.request.Protocol) == "https")
			if err != nil {
				fmt.Println("register client err:", err)
			}
		case <-s.ctx.Done():
			fmt.Println("ctx done")
			goto EXITLOOP
		}
	}
EXITLOOP:
}
