package serve

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/wrtx-dev/golocalsend/localsend/proto"
	"golang.org/x/net/ipv4"
)

const (
	MutlicastADDR = "224.0.0.167"
)

func ServeUDP(ctx context.Context) {
	l, err := net.ListenPacket("udp4", "0.0.0.0:53317")
	if err != nil {
		fmt.Println("listen udp err:", err)
		os.Exit(-1)
	}
	defer l.Close()
	conn := ipv4.NewPacketConn(l)
	err = conn.JoinGroup(nil, &net.UDPAddr{IP: net.ParseIP(MutlicastADDR)})
	if err != nil {
		fmt.Println("join mutlicast group err:", err)
		os.Exit(-1)
	}
	defer conn.LeaveGroup(nil, &net.UDPAddr{IP: net.ParseIP(MutlicastADDR)})
	defer conn.Close()

	flag := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("exited case ctx Done")
				flag <- struct{}{}
				return
			default:
				data := make([]byte, 4096)
				conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				n, _, addr, err := conn.ReadFrom(data)
				if err != nil {
					if e, ok := err.(net.Error); ok {
						if e.Timeout() {
							continue
						}
					}
					fmt.Println("read udp err:", err)
					continue
				}
				r, err := proto.ParseRegisterRequest(data[:n])
				if err != nil {
					fmt.Println("read mutlicast error:", err)
				}
				clientChan <- NewClientChan{
					request: r,
					addr:    addr,
				}
			}
		}
	}()
	<-flag
}
