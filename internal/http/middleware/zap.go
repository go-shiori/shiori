// Fiber middleware to enable zap logger for each request
// Adapted from https://gl.oddhunters.com/pub/fiberzap
package middleware

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ZapMiddlewareConfig defines the config for middleware
type ZapMiddlewareConfig struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c *fiber.Ctx) bool

	// Logger defines zap logger instance
	Logger *zap.Logger

	// CacheHeader defines the header name to get cache status from
	CacheHeader string
}

// New creates a new middleware handler
func NewZapMiddleware(config ZapMiddlewareConfig) fiber.Handler {
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

		fields := []zap.Field{
			zap.Namespace("context"),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status_code", c.Response().StatusCode()),
			zap.String("pid", strconv.Itoa(os.Getpid())),
			zap.String("time", stop.Sub(start).String()),
			zap.String("cache", string(c.Response().Header.Peek(config.CacheHeader))),
			zap.String("request-id", c.Locals("requestid").(string)),
		}

		formatErr := ""
		if chainErr != nil {
			formatErr = chainErr.Error()
			fields = append(fields, zap.String("error", formatErr))
			config.Logger.With(fields...).Error(formatErr)

			return nil
		}

		config.Logger.With(fields...).Info("request handled")

		return nil
	}
}
