package main

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// fetchCRL downloads and parses a CRL from the given URL.
func fetchCRL(url string) (*pkix.CertificateList, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch CRL: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Attempt to decode PEM if needed
	if p, _ := pem.Decode(data); p != nil {
		data = p.Bytes
	}
	crl, err := x509.ParseCRL(data)
	if err != nil {
		return nil, err
	}
	return crl, nil
}

// checkRevocation fetches the CRL partition based on serial prefix and checks if the serial is revoked.
func checkRevocation(serial, baseURL string) (bool, error) {
	serial = strings.TrimPrefix(strings.ToLower(serial), "0x")
	if len(serial) < 2 {
		return false, fmt.Errorf("serial too short")
	}
	prefix := serial[:2]
	url := fmt.Sprintf("%s/%s.crl", strings.TrimRight(baseURL, "/"), prefix)
	crl, err := fetchCRL(url)
	if err != nil {
		return false, err
	}
	trimmedSerial := strings.TrimLeft(serial, "0")
	for _, revoked := range crl.TBSCertList.RevokedCertificates {
		revokedHex := strings.TrimLeft(fmt.Sprintf("%x", revoked.SerialNumber), "0")
		if revokedHex == trimmedSerial {
			return true, nil
		}
	}
	return false, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <baseCRLURL> <serial>\n", os.Args[0])
		os.Exit(1)
	}
	baseURL := os.Args[1]
	serial := os.Args[2]
	revoked, err := checkRevocation(serial, baseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if revoked {
		fmt.Println("certificate is revoked")
	} else {
		fmt.Println("certificate is not revoked")
	}
}
