package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func generateSever() {
	caTpl, err := ReadCertificate("exampleCA.crt")
	if err != nil {
		logrus.Fatalln(err)
	}
	privateCaKey, err := ReadRSAPrivateKey("exampleCA.key")
	if err != nil {
		logrus.Fatalln(err)
	}

	privateSslKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logrus.Fatalln(err)
	}
	publicSslKey := privateSslKey.Public()
	subjectSsl := pkix.Name{
		CommonName:         "svr01",
		OrganizationalUnit: []string{"Example Org Unit"},
		Organization:       []string{"Example Org"},
		Country:            []string{"JP"},
	}
	before := time.Now().UTC()
	after := before.Add(24 * time.Hour * 365 * 10).UTC()
	sslTpl := &x509.Certificate{
		SerialNumber: big.NewInt(123),
		Subject:      subjectSsl,
		NotAfter:     before,
		NotBefore:    after,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}
	derSslCertificate, err := x509.CreateCertificate(rand.Reader, sslTpl, caTpl, publicSslKey, privateCaKey)
	if err != nil {
		logrus.Fatalln(err)
	}
	var f *os.File
	f, err = os.Create("svr01.crt")
	if err != nil {
		logrus.Fatalln(err)
	}
	err = pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: derSslCertificate})
	if err != nil {
		logrus.Fatalln(err)
	}
	err = f.Close()
	if err != nil {
		logrus.Fatalln(err)
	}

	f, err = os.Create("svr01.key")
	if err != nil {
		logrus.Fatalln(err)
	}
	derPrivateSslKey := x509.MarshalPKCS1PrivateKey(privateSslKey)
	err = pem.Encode(f, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: derPrivateSslKey})
	if err != nil {
		logrus.Fatalln(err)
	}
	err = f.Close()
	if err != nil {
		logrus.Fatalln(err)
	}
}
