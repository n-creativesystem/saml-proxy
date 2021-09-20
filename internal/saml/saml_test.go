package saml

import (
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/n-creativesystem/saml-proxy/internal/options"
	"github.com/n-creativesystem/saml-proxy/typ"
	"github.com/spf13/viper"
	"github.com/ucarion/urlpath"
)

var testConfig = `metadataUrl: http://example.com/metadata
x509Cert: ssl/server.crt
x509Key: ssl/server.key
redirect:
  rootURL: http://example.com/
  logoutUrl: /logout
upstream:
  url: http://example.com/proxy
  mapping:
    name: http://example.com/header.saml/claims
  resources:
    - url: /*
      methods:
        - "get"
      roles:
        - admin
  injectRequestHeaders:
    - name: X-Forwarded-xxx
      values:
        - claim: claimName
namedIdFormat: email
`

var expectedConfig = &Config{
	metadataURL: "http://example.com/metadata",
	x509Cert:    "ssl/server.crt",
	x509Key:     "ssl/server.key",
	redirect: struct {
		root   string
		logout string
	}{
		root:   "http://example.com/",
		logout: "/logout",
	},
	upstream: upstream{
		url: func() *url.URL {
			u, _ := url.Parse("http://example.com/proxy")
			return u
		}(),
		mapping: options.Mapping{
			Name:     "http://example.com/header.saml/claims",
			Multiple: false,
		},
		resources: urlPathResources{
			http.MethodGet: []pathRole{
				{
					path:  urlpath.New("/*"),
					roles: typ.StringSlice{"admin"}.ToMap(),
				},
			},
		},
		injectRequestHeaders: []options.Header{
			{
				Name: "X-Forwarded-xxx",
				Values: []options.HeaderValue{
					{
						ClaimSource: &options.ClaimSource{
							ClaimName: "claimName",
						},
					},
				},
			},
		},
	},
	nameIdFormat: "email",
	logoutURL:    "/saml/logout",
}

func TestSAMLConfig(t *testing.T) {
	const testFile = "test.yaml"
	file, _ := os.CreateTemp("", testFile)
	_, _ = file.WriteString(testConfig)
	file.Close()
	viper.Set("saml-config", file.Name())
	config := NewConfig()
	if !reflect.DeepEqual(expectedConfig, config) {
		t.Error("config deep equal")
	}
	os.Remove(testFile)
}
