package client

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/wrtx-dev/golocalsend/localsend/proto"
)

const (
	RegisterURI = "/api/localsend/v2/register"
)

func RegisterClient(r *proto.RegisterRequest, addr string, isHttps bool) error {
	b, err := proto.EncodeRegisterRequest(r)
	proto := "http"
	if isHttps {
		proto = "https"
	}
	if err != nil {
		return err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	hc := &http.Client{
		Transport: tr,
	}
	fmt.Println(fmt.Sprintf("%s://%s:53317%s", proto, addr, RegisterURI))
	resp, err := hc.Post(fmt.Sprintf("%s://%s:53317%s", proto, addr, RegisterURI), "application/json", bytes.NewReader(b))
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
		fmt.Println("Ignore error EOF")
	}
	if resp != nil {
		defer resp.Body.Close()
		fmt.Println("register status:", resp.Status)
	}
	return nil
}
