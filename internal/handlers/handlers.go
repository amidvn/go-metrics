package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/amidvn/go-metrics/internal/storage"
	"github.com/labstack/echo/v4"
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

		ctx.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
		return ctx.String(http.StatusOK, "")
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

func AllMetrics(s *storage.MemStorage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		err := ctx.String(http.StatusOK, s.AllMetrics())
		if err != nil {
			return err
		}

		return nil
	}
}
