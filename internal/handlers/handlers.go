package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amidvn/go-metrics/internal/models"
	"github.com/amidvn/go-metrics/internal/storage"
	"github.com/labstack/echo/v4"

	"go.uber.org/zap"
)

func PostWebhook(s *storage.MemStorage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		metricsType := ctx.Param("typeM")
		metricsName := ctx.Param("nameM")
		metricsValue := ctx.Param("valueM")

		switch metricsType {
		case "counter":
			value, err := strconv.ParseInt(metricsValue, 10, 64)
			if err != nil {
				return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s cannot be converted to an integer", metricsValue))
			}
			s.UpdateCounter(metricsName, value)
		case "gauge":
			value, err := strconv.ParseFloat(metricsValue, 64)
			if err != nil {
				return ctx.String(http.StatusBadRequest, fmt.Sprintf("%s cannot be converted to a float", metricsValue))
			}
			s.UpdateGauge(metricsName, value)
		default:
			return ctx.String(http.StatusBadRequest, "Invalid metric type. Can only be 'gauge' or 'counter'")
		}

		ctx.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return ctx.String(http.StatusOK, "")
	}
}

func UpdateJSON(s *storage.MemStorage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var metric models.Metrics
		err := json.NewDecoder(ctx.Request().Body).Decode(&metric)
		if err != nil {
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("Error in JSON decode: %s", err))
		}

		switch metric.MType {
		case "counter":
			s.UpdateCounter(metric.ID, *metric.Delta)
		case "gauge":
			s.UpdateGauge(metric.ID, *metric.Value)
		default:
			return ctx.String(http.StatusNotFound, "Invalid metric type. Can only be 'gauge' or 'counter'")
		}

		ctx.Response().Header().Set("Content-Type", "application/json")
		return ctx.JSON(http.StatusOK, metric)
	}
}

func MetricsValue(s *storage.MemStorage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		typeM := ctx.Param("typeM")
		nameM := ctx.Param("nameM")

		val, status := s.GetValue(typeM, nameM)
		err := ctx.String(status, val)
		if err != nil {
			return err
		}

		return nil
	}
}

func GetValueJSON(s *storage.MemStorage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var metric models.Metrics
		err := json.NewDecoder(ctx.Request().Body).Decode(&metric)
		if err != nil {
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("Error in JSON decode: %s", err))
		}

		switch metric.MType {
		case "counter":
			value := s.GetCounterValue(metric.ID)
			metric.Delta = &value
		case "gauge":
			value := s.GetGaugeValue(metric.ID)
			metric.Value = &value
		default:
			return ctx.String(http.StatusNotFound, "Invalid metric type. Can only be 'gauge' or 'counter'")
		}

		ctx.Response().Header().Set("Content-Type", "application/json")
		return ctx.JSON(http.StatusOK, metric)
	}
}

func AllMetrics(s *storage.MemStorage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set("Content-Type", "text/html")
		err := ctx.String(http.StatusOK, s.AllMetrics())
		if err != nil {
			return err
		}

		return nil
	}
}

func WithLogging(sugar zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			start := time.Now()

			req := ctx.Request()
			res := ctx.Response()
			if err = next(ctx); err != nil {
				ctx.Error(err)
			}
			duration := time.Since(start)

			sugar.Infoln(
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
