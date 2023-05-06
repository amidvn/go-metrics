package apiserver

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/amidvn/go-metrics/internal/storage"
)

type APIServer struct {
	storage *storage.MemStorage
}

func New() *APIServer {
	return &APIServer{storage.New()}
}

func (s *APIServer) Start() error {
	return http.ListenAndServe(`:8080`, http.HandlerFunc(s.webhook()))
}

func (s *APIServer) webhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		sliceURL := strings.Split(r.URL.Path, "/")

		if len(sliceURL) != 5 || sliceURL[1] != "update" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metricsType := sliceURL[2]
		metricsName := sliceURL[3]
		metricsValue := sliceURL[4]
		if metricsType == "counter" {
			if value, err := strconv.ParseInt(metricsValue, 10, 64); err == nil {
				s.storage.UpdateCounter(metricsName, value)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else if metricsType == "gauge" {
			if value, err := strconv.ParseFloat(metricsValue, 64); err == nil {
				s.storage.UpdateGauge(metricsName, value)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
}
