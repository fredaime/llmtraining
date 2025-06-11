package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// helper function to create a simple CA certificate and private key
func createCA(t *testing.T) (*x509.Certificate, crypto.Signer) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create cert: %v", err)
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("failed to parse cert: %v", err)
	}
	return cert, key
}

// helper to create a CRL including provided serials
func createCRL(t *testing.T, ca *x509.Certificate, key crypto.Signer, revokedSerials []string) []byte {
	var revokedList []pkix.RevokedCertificate
	for i, s := range revokedSerials {
		sn := new(big.Int)
		sn.SetString(s, 16)
		revokedList = append(revokedList, pkix.RevokedCertificate{
			SerialNumber:   sn,
			RevocationTime: time.Now().Add(time.Duration(i) * time.Minute),
		})
	}
	rlBytes, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{RevokedCertificates: revokedList, Number: big.NewInt(1), ThisUpdate: time.Now(), NextUpdate: time.Now().Add(1 * time.Hour)}, ca, key)
	if err != nil {
		t.Fatalf("failed to create CRL: %v", err)
	}
	return rlBytes
}

func TestFetchCRLSuccess(t *testing.T) {
	ca, key := createCA(t)
	crlDER := createCRL(t, ca, key, []string{"abcd"})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(crlDER)
	}))
	defer server.Close()

	crl, err := fetchCRL(server.URL)
	if err != nil {
		t.Fatalf("fetchCRL returned error: %v", err)
	}
	if len(crl.TBSCertList.RevokedCertificates) != 1 {
		t.Fatalf("expected 1 revoked cert, got %d", len(crl.TBSCertList.RevokedCertificates))
	}
}

func TestFetchCRLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	_, err := fetchCRL(server.URL)
	if err == nil {
		t.Fatalf("expected error from fetchCRL, got nil")
	}
}

func TestCheckRevocation(t *testing.T) {
	ca, key := createCA(t)
	crlDER := createCRL(t, ca, key, []string{"ab12"})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ab.crl" {
			w.Write(crlDER)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	revoked, err := checkRevocation("ab12", server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !revoked {
		t.Fatalf("expected serial to be revoked")
	}

	notRevoked, err := checkRevocation("ab34", server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notRevoked {
		t.Fatalf("expected serial to not be revoked")
	}
}

func TestCheckRevocationFetchError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	_, err := checkRevocation("ab12", server.URL)
	if err == nil {
		t.Fatalf("expected error but got none")
	}
}
