package server

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/n-creativesystem/saml-proxy/infra/redis"
	"github.com/n-creativesystem/saml-proxy/internal/saml"
	"github.com/n-creativesystem/saml-proxy/internal/search"
	"github.com/n-creativesystem/saml-proxy/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func New(opts ...Option) *gin.Engine {
	var samlMiddle *saml.Middleware
	conf := &config{}
	for _, opt := range opts {
		opt(conf)
	}
	con, err := redis.New(&redis.Config{
		Endpoints: []string{conf.redis},
	})
	if err != nil {
		logrus.Warnf("redis: %p", err)
	}
	if conf.debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DefaultWriter = &logger.HandlerLogger{
		Logger: logrus.StandardLogger(),
	}
	samlConfig := saml.NewConfig()
	samlMiddle = saml.New(samlConfig, con)
	p := ginprometheus.NewPrometheus(viper.GetString("metrics-name"))
	r := gin.New()
	r.Use(logger.RestLogger(), gin.Recovery())
	p.Use(r)
	r.Any("/saml/*actions", gin.WrapH(samlMiddle))
	if u := samlConfig.GetProxyURL(); u != nil {
		proxy := reverseProxy(samlConfig, samlMiddle.Session, con, u)
		r.Use(gin.WrapH(samlMiddle.RequireAccount(proxy)))
	} else {
		r.Use(func(c *gin.Context) {
			w := c.Writer
			r := c.Request
			session, err := samlMiddle.Session.GetSession(r)
			if err != nil {
				logrus.Errorf("Session error: %s", err)
			}
			if session != nil {
				buf, err := json.Marshal(session)
				if err != nil {
					logrus.Errorf("Session Json Marshal: %s", err)
				}
				src := map[string]interface{}{}
				dest := map[string]string{}
				_ = json.Unmarshal(buf, &src)
				// utils.MapFlatten(src, dest)
				strJwt := base64.RawURLEncoding.EncodeToString(buf)
				c.Request = r.WithContext(samlsp.ContextWithSession(r.Context(), session))
				header := w.Header()
				for key, value := range dest {
					header.Add(key, value)
				}
				w.Header().Add("x-saml-payload", strJwt)
				c.Status(http.StatusOK)
				return
			}
			if err == samlsp.ErrNoSession {
				samlMiddle.HandleStartAuthFlow(w, r)
				c.Abort()
				return
			}
			logrus.Error(err)
			c.AbortWithStatus(http.StatusUnauthorized)
		})
	}
	return r
}

func reverseProxy(samlConfig *saml.Config, sessionStore samlsp.SessionProvider, con redis.Redis, u *url.URL) http.Handler {
	t2 := http.DefaultTransport.(*http.Transport).Clone()
	if t2.TLSClientConfig == nil {
		t2.TLSClientConfig = &tls.Config{}
	}
	t2.TLSClientConfig.InsecureSkipVerify = samlConfig.InSecureSkipVerify()
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.FlushInterval = -1
	proxy.Transport = t2
	mapping := samlConfig.GetMapping()
	resources := samlConfig.GetResources()
	injectRequestHeader := samlConfig.GetInjectRequestHeaders()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var session samlsp.Session
		var err error
		if session, err = sessionStore.GetSession(r); err != nil {
			logrus.Errorf("Session error: %q", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if session != nil {
			buf, err := json.Marshal(session)
			if err != nil {
				logrus.Errorf("Session Json Marshal: %s", err)
			}
			src := map[string]interface{}{}
			_ = json.Unmarshal(buf, &src)
			search := search.New(src)
			var val []string
			if mapping.Multiple {
				val, err = search.GetSliceString(mapping.Name)
				if err != nil {
					logrus.Warn(err)
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(http.StatusText(http.StatusForbidden)))
					return
				}
			} else {
				v, err := search.GetString(mapping.Name)
				if err != nil {
					logrus.Warn(err)
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(http.StatusText(http.StatusForbidden)))
					return
				}
				val = []string{v}
			}
			if !resources.MultipleLookUp(r.Method, r.URL.Path, val) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(http.StatusText(http.StatusForbidden)))
				return
			}
			r = r.WithContext(samlsp.ContextWithSession(r.Context(), session))
			h := search.InjectHeader(injectRequestHeader, proxy)
			h.ServeHTTP(w, r)
			return
		} else {
			logrus.Errorf("Session error: %q", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	})
}
