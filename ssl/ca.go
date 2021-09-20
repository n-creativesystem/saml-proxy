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

func generateCA() {
	privateCaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		logrus.Fatalln(err)
	}
	publicCaKey := privateCaKey.Public()
	subjectCa := pkix.Name{
		CommonName:         "exampleCA",
		OrganizationalUnit: []string{"Example Org Unit"},
		Organization:       []string{"Example Org"},
		Country:            []string{"JP"},
	}
	before := time.Now().UTC()
	after := before.Add(24 * time.Hour * 365 * 10).UTC()
	caTpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               subjectCa,
		NotAfter:              after,
		NotBefore:             before,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	caCertificate, err := x509.CreateCertificate(rand.Reader, caTpl, caTpl, publicCaKey, privateCaKey)
	if err != nil {
		logrus.Fatalln(err)
	}

	var f *os.File
	f, err = os.Create("exampleCA.crt")
	if err != nil {
		logrus.Fatalln(err)
	}
	err = pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: caCertificate})
	if err != nil {
		logrus.Fatalln(err)
	}
	err = f.Close()
	if err != nil {
		logrus.Fatalln(err)
	}

	f, err = os.Create("exampleCA.key")
	if err != nil {
		logrus.Fatalln(err)
	}
	derCaPrivateKey := x509.MarshalPKCS1PrivateKey(privateCaKey)
	err = pem.Encode(f, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: derCaPrivateKey})
	if err != nil {
		logrus.Fatalln(err)
	}
	err = f.Close()
	if err != nil {
		logrus.Fatalln(err)
	}
}
