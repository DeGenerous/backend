package keys

import (
	"os"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

const keyPath = "./private_key.pem"

func GenerateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	pemPrivateFile, err := os.Create(keyPath)

	if err != nil {
		return nil, err
	}

	pemPrivateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	err = pem.Encode(pemPrivateFile, pemPrivateBlock)
	if err != nil {
		return nil, err
	}

	err = pemPrivateFile.Close()
	return privateKey, err
}

func ReadKey() (*rsa.PrivateKey, error) {
	bytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParseRawPrivateKey(bytes)
	if err != nil {
		return nil, err
	}

	return key.(*rsa.PrivateKey), nil
}
