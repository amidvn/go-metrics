package main

import (
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

type MemStorage struct {
	gaugeData   map[string]gauge
	counterData map[string]counter
}

var storage = MemStorage{
	gaugeData:   make(map[string]gauge),
	counterData: make(map[string]counter),
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	return http.ListenAndServe(`:8080`, http.HandlerFunc(webhook))
}

func webhook(w http.ResponseWriter, r *http.Request) {
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
			storage.counterData[metricsName] += counter(value)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else if metricsType == "gauge" {
		if value, err := strconv.ParseFloat(metricsValue, 64); err == nil {
			storage.gaugeData[metricsName] = gauge(value)
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
