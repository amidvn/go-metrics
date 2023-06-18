package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/amidvn/go-metrics/internal/database"
	"github.com/amidvn/go-metrics/internal/models"
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

func PingDB(db *database.DBConnection) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Header().Set("Content-Type", "text/html")
		err := database.CheckConnection(db)
		if err == nil {
			err = ctx.String(http.StatusOK, "Connection database is OK")
		} else {
			err = ctx.String(http.StatusInternalServerError, "Connection database is OK")
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func UpdatesJSON(s *storage.MemStorage) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		metrics := make([]models.Metrics, 0)
		err := json.NewDecoder(ctx.Request().Body).Decode(&metrics)
		fmt.Println("error", err)
		if err != nil && err != io.EOF {
			return ctx.String(http.StatusBadRequest, fmt.Sprintf("Error in JSON decode: %s", err))
		}

		s.StoreBatch(metrics)

		ctx.Response().Header().Set("Content-Type", "application/json")
		return ctx.JSON(http.StatusOK, metrics)
	}
}
