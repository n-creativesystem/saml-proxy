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
	cmd.Execute()
}
