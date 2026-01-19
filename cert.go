package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

// --- CERTIFICATE ---

func EnsureCert() {
	if _, err := os.Stat("server.crt"); os.IsNotExist(err) {
		fmt.Println("Generating self-signed certificate...")
		priv, _ := rsa.GenerateKey(rand.Reader, 2048)
		template := x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject:      pkix.Name{Organization: []string{"Freelancer Log Tool"}},
			NotBefore:    time.Now(),
			NotAfter:     time.Now().Add(365 * 24 * time.Hour),
			KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}
		derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
		fCert, _ := os.Create("server.crt")
		pem.Encode(fCert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		fCert.Close()
		fKey, _ := os.Create("server.key")
		pem.Encode(fKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
		fKey.Close()
	}
}
