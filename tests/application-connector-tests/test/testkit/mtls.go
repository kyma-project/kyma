package testkit

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

func NewMTLSClient(key *rsa.PrivateKey, certificates []*x509.Certificate) *http.Client {
	var rawCerts [][]byte

	for _, cert := range certificates {
		rawCerts = append(rawCerts, cert.Raw)
	}

	tlsCertificate := tls.Certificate{
		PrivateKey:  key,
		Certificate: rawCerts,
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{tlsCertificate},
			},
		},
	}
}
