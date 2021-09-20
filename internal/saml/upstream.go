package saml

import (
	"net/url"

	"github.com/n-creativesystem/saml-proxy/internal/options"
	"github.com/sirupsen/logrus"
)

type upstream struct {
	url                   *url.URL
	mapping               options.Mapping
	inSecureSkipTlsVerify bool
	resources             Matching
	injectRequestHeaders  []options.Header
}

func newUpstream(o options.Upstream) upstream {
	var u *url.URL
	var err error
	if proxyURL := o.URL; proxyURL != "" {
		u, err = url.Parse(proxyURL)
		if err != nil {
			logrus.Fatalln(err)
		}
	}
	baseResources := make(BaseResources, 0, len(o.Resources))
	for _, resource := range o.Resources {
		value := NewBaseResource(resource.URL, resource.Methods, resource.Roles)
		if value != nil {
			baseResources = append(baseResources, *value)
		}
	}
	upstream := upstream{
		url:                   u,
		mapping:               o.Mapping,
		inSecureSkipTlsVerify: o.InSecureSkipTlsVerify,
		resources:             baseResources.ToResources(),
		injectRequestHeaders:  options.Headers(o.InjectRequestHeaders).Clone(),
	}
	return upstream
}
