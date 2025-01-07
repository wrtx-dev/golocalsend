package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	r "golang.org/x/exp/rand"
)

type LocalsendConfig struct {
	HTTPS               bool   `json:"https"`
	SavePath            string `json:"savePath"`
	Port                uint   `json:"port"`
	MulticastAddr       string `json:"multicastAddr"`
	AutoSave            bool   `json:"autoSave"`
	Alias               string `json:"alias"`
	DeviceSearchTimeOut uint   `json:"deviceSearchTimeout"`
	DeviceModel         string `json:"deviceModel"`
	DeviceType          string `json:"deviceType"`
	RandStr             string
	Cert                *tls.Certificate
}

const (
	DefaultMutlicastADDR = "224.0.0.167"
)

func LoadConfigFile(file string) (*LocalsendConfig, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	buf, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	var localsendConfig = &LocalsendConfig{}
	err = json.Unmarshal(buf, localsendConfig)
	if err != nil {
		return nil, err
	}
	if localsendConfig.HTTPS {
		cert, sha, err := generateSelfSignedCert()
		if err != nil {
			return nil, err
		}
		localsendConfig.Cert = cert
		localsendConfig.RandStr = string(sha)
	}
	return localsendConfig, nil
}

func DefaultLocalsendConfig() (*LocalsendConfig, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cert, sha, err := generateSelfSignedCert()
	if err != nil {
		return nil, err
	}
	return &LocalsendConfig{
		HTTPS:               true,
		SavePath:            wd,
		Port:                53317,
		MulticastAddr:       DefaultMutlicastADDR,
		AutoSave:            false,
		Alias:               "golocalsend",
		DeviceSearchTimeOut: 500,
		DeviceModel:         "headless",
		DeviceType:          "headless",
		Cert:                cert,
		RandStr:             string(sha),
	}, nil

}

func generateSelfSignedCert() (*tls.Certificate, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"go localsend Self Signed Org"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * 10 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}
	sha := sha256.Sum256(certDER)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, err
	}
	return &cert, sha[:], nil
}

const (
	Letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = Letters[r.Int63()%int64(len(Letters))]
	}
	fmt.Println("rand string:", string(b))
	return string(b)
}
