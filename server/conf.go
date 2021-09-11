package server

import "github.com/sirupsen/logrus"

type config struct {
	samlConfig string
	debug      bool
	redis      string
	log        *logrus.Logger
}

type Option func(conf *config)

func WithSAMLConfig(filename string) Option {
	return func(conf *config) {
		conf.samlConfig = filename
	}
}

func WithDebug(conf *config) {
	conf.debug = true
}

func WithRedis(redis string) Option {
	return func(conf *config) {
		conf.redis = redis
	}
}

func WithLogger(log *logrus.Logger) Option {
	return func(conf *config) {
		conf.log = log
	}
}
