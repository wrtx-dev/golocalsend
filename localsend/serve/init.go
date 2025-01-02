package serve

import (
	"fmt"
	"net"

	"github.com/wrtx-dev/golocalsend/localsend/proto"
	"golang.org/x/exp/rand"
)

type NewClientChan struct {
	request *proto.RegisterRequest
	addr    net.Addr
}

var clientChan chan NewClientChan = nil

var MachineInfo *proto.RegisterRequest = nil

func init() {
	clientChan = make(chan NewClientChan)
	MachineInfo = &proto.RegisterRequest{
		Alias:        "golocalsend测试",
		Version:      "2.0",
		DeviceModel:  "headless",
		FingerPrint:  randString(65),
		Port:         53317,
		Protocol:     "http",
		Download:     false,
		Announcement: true,
		Announce:     true,
	}
}

const (
	Letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = Letters[rand.Int63()%int64(len(Letters))]
	}
	fmt.Println("rand string:", string(b))
	return string(b)
}
