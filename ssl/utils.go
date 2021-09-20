package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func ReadPem(filename string) (*pem.Block, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(buf)
	if block == nil {
		return nil, errors.New("invalid certificate data")
	}
	return block, nil
}

func ReadCertificate(filename string) (*x509.Certificate, error) {
	block, err := ReadPem(filename)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(block.Bytes)
}

func ReadRSAPrivateKey(filename string) (*rsa.PrivateKey, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(buf)
	if block == nil {
		return nil, errors.New("invalid private key data")
	}
	var key *rsa.PrivateKey
	if block.Type == "RSA PRIVATE KEY" {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	} else if block.Type == "PRIVATE KEY" {
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not RSA private key")
		}
	} else {
		return nil, fmt.Errorf("invalid private key type: %s", block.Type)
	}
	key.Precompute()
	if err := key.Validate(); err != nil {
		return nil, err
	}
	return key, nil
}
