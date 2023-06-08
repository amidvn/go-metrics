package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func WithLogging(logger zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			start := time.Now()

			req := ctx.Request()
			res := ctx.Response()
			if err = next(ctx); err != nil {
				ctx.Error(err)
			}
			duration := time.Since(start)

			logger.Infoln(
				"uri:", req.RequestURI,
				"method:", req.Method,
				"duration:", duration,
				"status:", res.Status,
				"size:", res.Size,
			)

			return err
		}
	}
}

func GzipUnpacking() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			req := ctx.Request()
			rw := ctx.Response().Writer
			header := req.Header
			if strings.Contains(header.Get("Accept-Encoding"), "gzip") {
				cw := newCompressWriter(rw)
				ctx.Response().Writer = cw
				defer cw.Close()
			}

			if strings.Contains(header.Get("Content-Encoding"), "gzip") {
				cr, err := newCompressReader(req.Body)
				if err != nil {
					return ctx.String(http.StatusInternalServerError, "")
				}
				ctx.Request().Body = cr
				defer cr.Close()
			}
			if err = next(ctx); err != nil {
				ctx.Error(err)
			}

			return err
		}
	}
}
