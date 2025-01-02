package serve

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/wrtx-dev/golocalsend/localsend/proto"
)

func preupload(w http.ResponseWriter, req *http.Request) {

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "无法读取请求 body", http.StatusInternalServerError)
		return
	}
	r, err := proto.ParsePreUploadRequest(body)
	if err != nil {
		fmt.Println("parse", string(body), "err:", err)
	}
	fmt.Printf("r:%+v\n", r)
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("403 Forbidden"))

}

func HandleFileServer(ctx context.Context) {
	http.HandleFunc("/api/localsend/v2/prepare-upload", preupload)
	server := &http.Server{
		Addr:    ":53317",
		Handler: nil,
	}
	hctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务器启动失败: %v", err)
		}
	}()
	<-ctx.Done()
	server.Shutdown(hctx)
}
