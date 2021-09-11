package main

import (
	"github.com/n-creativesystem/saml-proxy/cmd"
	"github.com/n-creativesystem/saml-proxy/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: logger.TimestampFormat,
	})
	if err := cmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
}
