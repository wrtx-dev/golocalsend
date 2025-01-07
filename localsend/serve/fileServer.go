package serve

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/wrtx-dev/golocalsend/localsend/proto"
)

const (
	PENNDING = iota
	UPLOADING
	CANCELED
)

type FComfirmUpload func(map[string]proto.FileInfo) (map[string]proto.FileInfo, error)

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
	preupReqChan <- *r
	resp := <-preupRespChan
	if len(resp.Files) == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	respBuf, err := proto.EncodePreUploadResponse(&resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(respBuf)

}

func upload(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	fileCh := make(chan fileInfo)
	defer close(fileCh)
	qi := queryInfo{
		sessionID: query.Get("sessionId"),
		fileId:    query.Get("fileId"),
		token:     query.Get("token"),
		ch:        fileCh,
	}
	queryChan <- qi
	info := <-fileCh
	if info.errMsg != "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	defer func() {
		finshChan <- info.id
	}()
	absPath, err := filepath.Abs(info.savePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fp, err := os.Create(absPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer fp.Close()
	b := make([]byte, 4096)
	for {
		n, err := req.Body.Read(b)
		if err != nil {
			if !errors.Is(io.EOF, err) {
				fmt.Println("recv error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		fp.Write(b[:n])
		if err != nil && errors.Is(io.EOF, err) {
			break
		}
	}
	w.WriteHeader(http.StatusOK)
	return
}

func cancelUpload(w http.ResponseWriter, req *http.Request) {
	sessionId := req.URL.Query().Get("sessionId")
	if sessionId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cancelChan <- sessionId
	w.WriteHeader(http.StatusOK)
}

func RegisterRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rr, err := proto.ParseRegisterRequest(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	addr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	clientChan <- NewClientChan{
		request: rr,
		addr:    addr,
	}
	w.WriteHeader(http.StatusOK)
}

func (s *GoLocalsendServer) HandleFileServer() {

	http.HandleFunc("/api/localsend/v2/prepare-upload", preupload)
	http.HandleFunc("/api/localsend/v2/upload", upload)
	http.HandleFunc("/api/localsend/v2/cancel", cancelUpload)
	http.HandleFunc("/api/localsend/v2/register", RegisterRequest)
	var server *http.Server

	server = &http.Server{
		Addr:    ":53317",
		Handler: nil,
	}
	cert := s.config.Cert
	if cert != nil {
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{*cert},
		}
	}
	hctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		serve := func() error {
			if cert != nil {
				return server.ListenAndServeTLS("", "")
			} else {
				return server.ListenAndServe()
			}
		}
		if err := serve(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务器启动失败: %v", err)
		}
	}()
	<-s.ctx.Done()
	server.Shutdown(hctx)
}

type fileInfo struct {
	id        string
	name      string
	size      uint64
	info      proto.Info
	token     string
	sessionId string
	errMsg    string
	savePath  string
	status    int
}

type queryInfo struct {
	sessionID string
	fileId    string
	token     string
	ch        chan<- fileInfo
}

var preupReqChan chan proto.PreUploadRequest = nil
var preupRespChan chan proto.PreUploadResponse = nil
var queryChan chan queryInfo = nil
var finshChan chan string = nil
var cancelChan chan string = nil

func (s *GoLocalsendServer) fileRecorder() {
	preupReqChan = make(chan proto.PreUploadRequest)
	defer close(preupReqChan)
	preupRespChan = make(chan proto.PreUploadResponse)
	defer close(preupRespChan)
	queryChan = make(chan queryInfo)
	defer close(queryChan)
	finshChan = make(chan string)
	defer close(finshChan)
	cancelChan = make(chan string)
	defer close(cancelChan)
	fileMap := make(map[string]fileInfo)
	for {
		select {
		case <-s.ctx.Done():
			return
		case req := <-preupReqChan:
			uuidv4 := uuid.New()
			resp := proto.PreUploadResponse{
				SessionID: uuidv4.String(),
				Files:     map[string]string{},
			}
			files := req.Files
			if s.comfirm != nil {
				files, _ = s.comfirm(req.Files)
			}
			for _, v := range files {
				if _, ok := fileMap[v.ID]; ok {
					break
				}
				token := uuid.New().String()
				fileMap[v.ID] = fileInfo{
					id:        v.ID,
					name:      v.FileName,
					size:      v.Size,
					info:      req.Info,
					token:     token,
					sessionId: uuidv4.String(),
					savePath:  s.getSavePath(v.FileName),
					status:    UPLOADING,
				}
				resp.Files[v.ID] = token
			}
			preupRespChan <- resp
		case query := <-queryChan:
			if q, ok := fileMap[query.fileId]; ok {
				query.ch <- q
			} else {
				query.ch <- fileInfo{
					errMsg: "not found",
				}
			}
		case token := <-finshChan:
			if f, ok := fileMap[token]; ok {
				if f.status == CANCELED {
					err := os.Remove(f.savePath)
					if err != nil {
						fmt.Println("remove file:", f.savePath, "err:", err)
					}
				}
			}
			delete(fileMap, token)
		case sessionId := <-cancelChan:
			for k, v := range fileMap {
				if v.sessionId == sessionId {
					if v.status == UPLOADING {
						v.status = CANCELED
						fileMap[k] = v
					} else {
						delete(fileMap, k)
					}

				}
			}
		}
	}
}

func (s *GoLocalsendServer) getSavePath(name string) string {
	return filepath.Join(s.config.SavePath, name)
}
