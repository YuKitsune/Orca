package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

const blockType = "RSA PRIVATE KEY"

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
	if block == nil || block.Type != blockType {
		return nil, errors.New("failed to decode PEM block containing RSA private key")
	}

	// TODO: Is this the right decode function?
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func EncodePrivateKey(privateKey *rsa.PrivateKey) []byte {
	bytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  blockType,
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	return bytes
}
