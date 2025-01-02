package serve

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/wrtx-dev/golocalsend/localsend/client"
)

func RegisterHTTPClient(ctx context.Context) {
	for {
		select {
		case req := <-clientChan:
			fmt.Println(req.request)
			err := client.RegisterClient(MachineInfo, req.addr.(*net.UDPAddr).IP.String(), strings.ToLower(req.request.Protocol) == "https")
			if err != nil {
				fmt.Println("register client err:", err)
			}
		case <-ctx.Done():
			fmt.Println("ctx done")
			goto EXITLOOP
		}
	}
EXITLOOP:
}
