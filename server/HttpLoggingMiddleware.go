package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/sdks-service/logger"
	"go.uber.org/zap"
)

type LoggingConfig struct {
	Fields             []string
	ServerErrorMessage string
	ClientErrorMessage string
	SuccessMessage     string
}

func NewHttpLoggingMiddleware(config *LoggingConfig) fiber.Handler {

	if len(config.Fields) == 0 {
		config.Fields = []string{
			"ip",
			"host",
			"path",
			"statusCode",
			"latency",
			"bytesSent",
			"bytesReceived",
			"requestId",
			"error",
		}
	}

	if config.ServerErrorMessage == "" {
		config.ServerErrorMessage = "Server error"
	}

	if config.ClientErrorMessage == "" {
		config.ClientErrorMessage = "Client error"
	}

	if config.SuccessMessage == "" {
		config.SuccessMessage = "Request Processed Successfully"
	}

	return func(c *fiber.Ctx) (err error) {
		var latency int64

		start := time.Now()
		chainErr := c.Next()
		// we must call the error handler or we get a 200 status code in the logs -_-
		if chainErr != nil {
			c.App().ErrorHandler(c, chainErr)
		}
		latency = time.Since(start).Milliseconds()

		fields := map[string]interface{}{}
		statusCode := c.Response().StatusCode()

		fields["statusCode"] = statusCode
		fields["ip"] = c.IP()

		for _, field := range config.Fields {
			switch field {
			case "referer":
				fields["referer"] = c.Get(fiber.HeaderReferer)
			case "protocol":
				fields["protocol"] = c.Protocol()
			case "port":
				fields["port"] = c.Port()
			case "ip":
				fields["ip"] = c.IP()
			case "ips":
				fields["ips"] = c.Get(fiber.HeaderXForwardedFor)
			case "host":
				fields["host"] = c.Hostname()
			case "path":
				fields["path"] = c.Path()
			case "url":
				fields["url"] = c.OriginalURL()
			case "ua":
				fields["ua"] = c.Get(fiber.HeaderUserAgent)
			case "latency":
				fields["latency"] = latency
			case "statusCode":
				fields["statusCode"] = statusCode
			case "queryParams":
				fields["queryParams"] = c.Request().URI().QueryArgs().String()
			case "bytesReceived":
				fields["bytesReceived"] = len(c.Request().Body())
			case "bytesSent":
				fields["bytesSent"] = len(c.Response().Body())
			case "route":
				fields["route"] = c.Route().Path
			case "method":
				fields["method"] = c.Method()
			case "requestId":
				rid := c.GetRespHeader(fiber.HeaderXRequestID)
				if rid != "" {
					fields["requestId"] = rid
				}
			}
		}

		log := logger.Logger.WithOptions(zap.WithCaller(false)).With(zap.Any("http", fields))

		if chainErr != nil {
			log = log.With(zap.Error(chainErr))
		}

		// Return fields by status code
		switch {
		case statusCode >= 500:
			log.Error(config.ServerErrorMessage)
		case statusCode >= 400:
			log.Warn(config.ClientErrorMessage)
		default:
			log.Info(config.SuccessMessage)
		}

		return nil
	}
}
