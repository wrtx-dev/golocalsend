package proto

import "encoding/json"

type RegisterRequest struct {
	Alias        string `json:"alias"`
	Version      string `json:"version"`
	DeviceModel  string `json:"deviceModel"`
	FingerPrint  string `json:"fingerprint"`
	Port         int    `json:"port"`
	Protocol     string `json:"protocol"`
	Download     bool   `json:"download"`
	Announcement bool   `json:"announcement"`
	Announce     bool   `json:"announce"`
}

type Info struct {
	Alias       string `json:"alias"`
	Version     string `json:"version"`
	DeviceModel string `json:"deviceModel"`
	FingerPrint string `json:"fingerprint"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Download    bool   `json:"download"`
}

type FileInfo struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	Size     int64  `json:"size"`
	FileType string `json:"fileType"`
	Sha256   string `json:"sha256"`
	Preview  string `json:"preview"`
	Metadata struct {
		Modified string `json:"modified"`
	} `json:"metadata"`
}

type PreUploadRequest struct {
	Info  Info                `json:"info"`
	Files map[string]FileInfo `json:"files"`
}

func ParseRegisterRequest(buf []byte) (*RegisterRequest, error) {
	r := &RegisterRequest{}
	err := json.Unmarshal(buf, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func EncodeRegisterRequest(r *RegisterRequest) ([]byte, error) {
	b, err := json.Marshal(r)
	return b, err
}

func ParsePreUploadRequest(b []byte) (*PreUploadRequest, error) {
	r := &PreUploadRequest{}
	err := json.Unmarshal(b, r)
	return r, err
}
