// https://golang.org/src/crypto/tls/generate_cert.go

package gypsum

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func getTlsKeys() (publicKeyPath, privateKeyPath string, err error) {
	publicKeyPath = "./gypsum.pem"
	privateKeyPath = "./gypsum.key"
	_, e := os.Stat(publicKeyPath)
	if e != nil {
		if os.IsNotExist(e) {
			err = newSelfSignedKeys(publicKeyPath, privateKeyPath)
			return
		}
		err = e
		return
	}
	_, e = os.Stat(privateKeyPath)
	if e != nil {
		if os.IsNotExist(e) {
			err = newSelfSignedKeys(publicKeyPath, privateKeyPath)
			return
		}
		err = e
		return
	}
	return
}

func newSelfSignedKeys(publicKeyPath, privateKeyPath string) error {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"gypsum"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 10),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk)
	if err != nil {
		return err
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return err
	}
	pemFile, _ := os.Create(publicKeyPath)
	defer pemFile.Close()
	err = pem.Encode(pemFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return err
	}
	keyFile, _ := os.Create(privateKeyPath)
	defer keyFile.Close()
	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	if err != nil {
		return err
	}
	return nil
}
