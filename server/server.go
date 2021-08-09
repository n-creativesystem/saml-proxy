package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/n-creativesystem/saml-proxy/infra/redis"
	"github.com/n-creativesystem/saml-proxy/internal/saml"
	"github.com/n-creativesystem/saml-proxy/logger"
	"github.com/sirupsen/logrus"
)

func New(opts ...Options) *gin.Engine {
	log := logger.NewHandlerLogger()
	var samlMiddle *samlsp.Middleware
	conf := &config{}
	for _, opt := range opts {
		opt(conf)
	}
	con, err := redis.New(&redis.Config{
		Endpoints: []string{conf.redis},
	})
	if err != nil {
		log.Warnf("redis: %v", err)
	}
	loggerOpts := []logger.HandlerLogOption{}
	if conf.debug {
		log.SetLevel(logrus.DebugLevel)
		loggerOpts = append(loggerOpts, logger.WithGinDebug(logrus.DebugLevel))
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
		log.SetLevel(logrus.InfoLevel)
	}
	gin.DefaultWriter = log
	r := gin.New()
	r.Use(logger.RestLogger(loggerOpts...), gin.Recovery())
	samlMiddle = saml.New(saml.WithFilename(conf.samlConfig), saml.WithRedis(con))
	r.Any("/saml/*action", func(c *gin.Context) {
		w := c.Writer
		r := c.Request
		samlMiddle.ServeHTTP(w, r)
	})
	r.Use(func(c *gin.Context) {
		w := c.Writer
		r := c.Request
		session, err := samlMiddle.Session.GetSession(r)
		if session != nil {
			buf, _ := json.Marshal(session)
			strJwt := base64.RawURLEncoding.EncodeToString(buf)
			c.Request = r.WithContext(samlsp.ContextWithSession(r.Context(), session))
			w.Header().Add("x-saml-payload", strJwt)
			c.Status(http.StatusOK)
			return
		}
		if err == samlsp.ErrNoSession {
			samlMiddle.HandleStartAuthFlow(w, r)
			c.Abort()
			return
		}
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
	})

	return r
}
