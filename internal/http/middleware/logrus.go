// Fiber middleware to enable zap logger for each request
// Adapted from https://gl.oddhunters.com/pub/fiberzap
package middleware

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// LogrusMiddlewareConfig defines the config for middleware
type LogrusMiddlewareConfig struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c *fiber.Ctx) bool

	// Logger defines logrus logger instance
	Logger *logrus.Logger

	// CacheHeader defines the header name to get cache status from
	CacheHeader string
}

// New creates a new middleware handler
func NewLogrusMiddleware(config LogrusMiddlewareConfig) fiber.Handler {
	var (
		errPadding  = 15
		start, stop time.Time
		once        sync.Once
		errHandler  fiber.ErrorHandler
	)

	return func(c *fiber.Ctx) error {
		if config.Next != nil && config.Next(c) {
			return c.Next()
		}

		once.Do(func() {
			errHandler = c.App().Config().ErrorHandler
			stack := c.App().Stack()
			for m := range stack {
				for r := range stack[m] {
					if len(stack[m][r].Path) > errPadding {
						errPadding = len(stack[m][r].Path)
					}
				}
			}
		})

		start = time.Now()

		chainErr := c.Next()

		if chainErr != nil {
			if err := errHandler(c, chainErr); err != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		stop = time.Now()

		fields := logrus.Fields{
			"method":      c.Method(),
			"path":        c.Path(),
			"status_code": c.Response().StatusCode(),
			"pid":         strconv.Itoa(os.Getpid()),
			"duration":    stop.Sub(start).String(),
			"cache":       string(c.Response().Header.Peek(config.CacheHeader)),
			"request-id":  c.Locals("requestid").(string),
		}
		l := config.Logger.WithFields(fields)

		if chainErr != nil {
			l = l.WithError(chainErr)
		}

		msg := c.Method() + " " + string(c.Context().RequestURI())
		if c.Response().StatusCode() == fiber.StatusOK {
			l.Info(msg)
		} else {
			l.Warn(msg)
		}

		return nil
	}
}
