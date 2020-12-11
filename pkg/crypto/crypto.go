package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

func DecodePrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {

	// Read the file
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decide the key
	key, err := DecodePrivateKey(raw)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func DecodePrivateKey(raw []byte) (*rsa.PrivateKey, error) {

	// Decode
	block, _ := pem.Decode(raw)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing RSA private key")
	}

	// TODO: Is this the right decode function?
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}