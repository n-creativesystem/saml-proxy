package saml

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"os"
	"path"

	"github.com/crewjam/saml/samlsp"
	"github.com/n-creativesystem/saml-proxy/infra/redis"
	"github.com/n-creativesystem/saml-proxy/internal/render"
)

type config struct {
	filename string
	con      redis.Redis
}

type Options func(conf *config)

func WithFilename(filename string) Options {
	return func(conf *config) {
		conf.filename = filename
	}
}

func WithRedis(con redis.Redis) Options {
	return func(conf *config) {
		conf.con = con
	}
}

type SAMLConfig struct {
	MetadataURL string `yaml:"metadata_url" json:"metadata_url"`
	X509Cert    string `yaml:"x509_cert" json:"x509_cert"`
	X509Key     string `yaml:"x509_key" json:"x509_key"`
	RootURL     string `yaml:"root_url" json:"root_url"`
	IDPLogout   string `yaml:"idp_logout" json:"idp_logout"`
}

var SAMLConf SAMLConfig

var defaultConfig = config{
	filename: "saml.yaml",
}

func New(opts ...Options) *samlsp.Middleware {
	conf := &config{}
	*conf = defaultConfig
	for _, opt := range opts {
		opt(conf)
	}
	filename := conf.filename
	buf, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	render := filenameIsPkg(filename)
	err = render.Unmarshal(buf, &SAMLConf)
	if err != nil {
		panic(err)
	}
	sp := ServiceProvider(conf)
	return sp
}

func filenameIsPkg(filename string) render.Render {
	switch path.Ext(filename) {
	case ".yaml", ".yml":
		return render.YAML
	case ".json":
		return render.JSON
	default:
		return nil
	}
}

func ServiceProvider(conf *config) *samlsp.Middleware {
	type certificate struct {
		privateKey  *rsa.PrivateKey
		certificate *x509.Certificate
	}
	var cert certificate
	if SAMLConf.X509Cert != "" && SAMLConf.X509Key != "" {
		keyPair, err := tls.LoadX509KeyPair(SAMLConf.X509Cert, SAMLConf.X509Key)
		if err != nil {
			panic(err)
		}
		keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
		if err != nil {
			panic(err)
		}
		cert.privateKey = keyPair.PrivateKey.(*rsa.PrivateKey)
		cert.certificate = keyPair.Leaf
	}
	idpMetadataURL, err := url.Parse(SAMLConf.MetadataURL)
	if err != nil {
		panic(err)
	}
	rootURL, err := url.Parse(SAMLConf.RootURL)
	if err != nil {
		panic(err)
	}

	opts := samlsp.Options{
		URL:            *rootURL,
		Key:            cert.privateKey,
		Certificate:    cert.certificate,
		IDPMetadataURL: idpMetadataURL,
	}
	samlSP, err := samlsp.New(opts)
	samlSP.Session = newSessionProvider(opts, conf.con)
	if err != nil {
		panic(err)
	}
	return samlSP
}
