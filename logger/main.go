package logger

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	TimestampFormat = "2006/01/02 - 15:04:05"
)

type handlerLogConfig struct {
	logLevel logrus.Level
}

type HandlerLogOption func(conf *handlerLogConfig)

func WithGinDebug(level logrus.Level) HandlerLogOption {
	return func(conf *handlerLogConfig) {
		conf.logLevel = level
	}
}

type handlerLogger struct {
	*logrus.Logger
}

var _ io.Writer = (*handlerLogger)(nil)

var logPool *sync.Pool

func init() {
	logPool = &sync.Pool{
		New: func() interface{} {
			log := logrus.New()
			log.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: TimestampFormat,
			})
			return &handlerLogger{
				Logger: log,
			}
		},
	}
}

func NewHandlerLogger() *handlerLogger {
	return logPool.Get().(*handlerLogger)
}

func (l *handlerLogger) Write(p []byte) (n int, err error) {
	l.Logger.Debug(string(p))
	return len(p), nil
}

func RestLogger(opts ...HandlerLogOption) gin.HandlerFunc {
	conf := &handlerLogConfig{
		logLevel: logrus.InfoLevel,
	}
	for _, opt := range opts {
		opt(conf)
	}
	return func(c *gin.Context) {
		log := NewHandlerLogger()
		log.SetLevel(conf.logLevel)

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		ctx := c.Request.Context()
		mpHeader := c.Request.Header.Clone()
		newCtx := ToContext(ctx, log)
		*c.Request = *c.Request.WithContext(newCtx)
		c.Next()

		param := gin.LogFormatterParams{
			Request: c.Request,
			Keys:    c.Keys,
		}

		param.TimeStamp = time.Now()
		param.Latency = param.TimeStamp.Sub(start)

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

		param.BodySize = c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		param.Path = path
		fields := logrus.Fields{
			"key":      "SAML-PROXY",
			"status":   param.StatusCode,
			"latency":  param.Latency,
			"clientIP": param.ClientIP,
			"method":   param.Method,
			"path":     param.Path,
			"Ua":       param.Request.UserAgent(),
		}
		if conf.logLevel == logrus.DebugLevel {
			for key, value := range mpHeader {
				if len(value) >= 0 {
					log.Debugf("req_%s: %v", key, value)
				}
			}
		}
		if conf.logLevel == logrus.DebugLevel {
			mpHeader := c.Writer.Header().Clone()
			for key, value := range mpHeader {
				if len(value) >= 0 {
					log.Debugf("res_%s: %v", key, value)
				}
			}
		}
		log.WithFields(fields).Info("incoming request")
		logPool.Put(log)
	}
}

type ctxLoggerMarker struct{}

var logKey = &ctxLoggerMarker{}

func FromContext(ctx context.Context) *handlerLogger {
	if val, ok := ctx.Value(logKey).(*handlerLogger); ok && val != nil {
		return val
	}
	log := NewHandlerLogger()
	log.SetLevel(logrus.DebugLevel)
	return log
}

func ToContext(ctx context.Context, log *handlerLogger) context.Context {
	return context.WithValue(ctx, logKey, log)
}
