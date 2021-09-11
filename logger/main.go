package logger

import (
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	TimestampFormat = "2006/01/02 - 15:04:05"
)

type HandlerLogger struct {
	*logrus.Logger
}

var _ io.Writer = (*HandlerLogger)(nil)

func (l *HandlerLogger) Write(p []byte) (n int, err error) {
	l.Logger.Debug(string(p))
	return len(p), nil
}

func RestLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		ctx := c.Request.Context()
		mpHeader := c.Request.Header.Clone()
		for key, value := range mpHeader {
			if len(value) >= 0 {
				log.Debugf("req_%s: %v", key, value)
			}
		}
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
		mpHeader = c.Writer.Header().Clone()
		for key, value := range mpHeader {
			if len(value) >= 0 {
				log.Debugf("res_%s: %v", key, value)
			}
		}
		log.WithFields(fields).Info("incoming request")
	}
}

type ctxLoggerMarker struct{}

var logKey = &ctxLoggerMarker{}

func FromContext(ctx context.Context) *logrus.Logger {
	if val, ok := ctx.Value(logKey).(*logrus.Logger); ok && val != nil {
		return val
	}
	return nil
}

func ToContext(ctx context.Context, log *logrus.Logger) context.Context {
	return context.WithValue(ctx, logKey, log)
}
