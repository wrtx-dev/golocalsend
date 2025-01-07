package serve

var clientChan chan NewClientChan = nil

func init() {
	clientChan = make(chan NewClientChan)

}
