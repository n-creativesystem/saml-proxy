package saml

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"path"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/n-creativesystem/saml-proxy/infra/redis"
	"github.com/n-creativesystem/saml-proxy/internal/options"
	"github.com/n-creativesystem/saml-proxy/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const defaultLogoutURL = "/logout"

type viperConfig struct {
	MetadataURL string           `json:"metadataUrl"`
	X509Cert    string           `json:"x509Cert"`
	X509Key     string           `json:"x509Key"`
	Redirect    options.Redirect `json:"redirect"`
	// IdpLogout     string           `json:"idpLogout"`
	CookieName    string           `json:"cookieName"`
	Upstream      options.Upstream `json:"upstream"`
	NamedIdFormat string           `json:"namedIdFormat"`
	LogoutURL     string           `json:"logoutUrl"`
}

func newViperConfig() (*viperConfig, error) {
	baseConfig := viperConfig{}
	if err := utils.LoadYAML(viper.GetString("saml-config"), &baseConfig); err != nil {
		return nil, err
	}
	return &baseConfig, nil
}

type redirect struct {
	root   string
	logout string
}

type Config struct {
	metadataURL string
	x509Cert    string
	x509Key     string
	redirect    redirect
	// idpLogout    string
	cookieName   string
	nameIdFormat string
	logoutURL    string
	upstream     upstream
}

func (c *Config) GetProxyURL() *url.URL {
	return c.upstream.url
}

func (c *Config) GetResources() Matching {
	return c.upstream.resources
}

func (c *Config) GetMapping() options.Mapping {
	return c.upstream.mapping
}

func (c *Config) InSecureSkipVerify() bool {
	return !c.upstream.inSecureSkipTlsVerify
}

func (c *Config) GetInjectRequestHeaders() []options.Header {
	return c.upstream.injectRequestHeaders
}

func NewConfig() *Config {
	baseConfig, err := newViperConfig()
	if err != nil {
		logrus.Fatalln(err)
	}
	redirectLogout := "/"
	if baseConfig.Redirect.LogoutRedirect != "" {
		u, err := url.Parse(baseConfig.Redirect.LogoutRedirect)
		if err != nil {
			logrus.Fatalln(err)
		}
		redirectLogout = u.RequestURI()
	}
	if baseConfig.LogoutURL == "" {
		baseConfig.LogoutURL = defaultLogoutURL
	}
	logoutURL := path.Join("/saml", baseConfig.LogoutURL)
	config := &Config{
		metadataURL: baseConfig.MetadataURL,
		x509Cert:    baseConfig.X509Cert,
		x509Key:     baseConfig.X509Key,
		redirect: redirect{
			root:   baseConfig.Redirect.RootURL,
			logout: redirectLogout,
		},
		cookieName:   baseConfig.CookieName,
		nameIdFormat: baseConfig.NamedIdFormat,
		logoutURL:    logoutURL,
		upstream:     newUpstream(baseConfig.Upstream),
	}
	return config
}

type Middleware struct {
	*samlsp.Middleware
	config *Config
}

var _ http.Handler = (*Middleware)(nil)

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == m.config.logoutURL {
		if err := m.Middleware.Session.DeleteSession(w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			http.Redirect(w, r, m.config.redirect.logout, http.StatusFound)
		}
	} else {
		m.Middleware.ServeHTTP(w, r)
	}
}

func New(config *Config, con redis.Redis) *Middleware {
	sp := ServiceProvider(config, con)
	return &Middleware{
		Middleware: sp,
		config:     config,
	}
}

func ServiceProvider(config *Config, con redis.Redis) *samlsp.Middleware {
	type certificate struct {
		privateKey  *rsa.PrivateKey
		certificate *x509.Certificate
	}
	var cert certificate
	if config.x509Cert != "" && config.x509Key != "" {
		keyPair, err := tls.LoadX509KeyPair(config.x509Cert, config.x509Key)
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
	idpMetadataURL, err := url.Parse(config.metadataURL)
	if err != nil {
		panic(err)
	}
	loginRedirect, err := url.Parse(config.redirect.root)
	if err != nil {
		panic(err)
	}

	opts := samlsp.Options{
		URL:            *loginRedirect,
		Key:            cert.privateKey,
		Certificate:    cert.certificate,
		IDPMetadataURL: idpMetadataURL,
	}
	samlSP, err := samlsp.New(opts)
	samlSP.Session = NewSessionProvider(config, opts, con)
	if err != nil {
		panic(err)
	}
	switch config.nameIdFormat {
	case "unspecified":
		samlSP.ServiceProvider.AuthnNameIDFormat = saml.UnspecifiedNameIDFormat
	case "transient":
		samlSP.ServiceProvider.AuthnNameIDFormat = saml.TransientNameIDFormat
	case "email":
		samlSP.ServiceProvider.AuthnNameIDFormat = saml.EmailAddressNameIDFormat
	case "persistent":
		samlSP.ServiceProvider.AuthnNameIDFormat = saml.PersistentNameIDFormat
	default:
		samlSP.ServiceProvider.AuthnNameIDFormat = saml.NameIDFormat(config.nameIdFormat)
	}
	return samlSP
}
