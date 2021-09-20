package server

type config struct {
	samlConfig string
	debug      bool
	redis      string
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
