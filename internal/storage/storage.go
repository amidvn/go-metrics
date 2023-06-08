package storage

import (
	"fmt"
	"net/http"
)

type gauge float64
type counter int64

type MemStorage struct {
	gaugeData   map[string]gauge
	counterData map[string]counter
}

type AllMetrics struct {
	Gauge   map[string]gauge   `json:"gauge"`
	Counter map[string]counter `json:"counter"`
}

func New(storeInterval int, filePath string, restore bool) *MemStorage {
	storage := MemStorage{
		gaugeData:   make(map[string]gauge),
		counterData: make(map[string]counter),
	}

	return &storage
}

func (s *MemStorage) UpdateCounter(n string, v int64) {
	s.counterData[n] += counter(v)
}

func (s *MemStorage) UpdateGauge(n string, v float64) {
	s.gaugeData[n] = gauge(v)
}

func (s *MemStorage) GetValue(t string, n string) (string, int) {
	var v string
	statusCode := http.StatusOK
	if val, ok := s.gaugeData[n]; ok && t == "gauge" {
		v = fmt.Sprint(val)
	} else if val, ok := s.counterData[n]; ok && t == "counter" {
		v = fmt.Sprint(val)
	} else {
		statusCode = http.StatusNotFound
	}
	return v, statusCode
}

func (s *MemStorage) GetCounterValue(id string) int64 {
	return int64(s.counterData[id])
}

func (s *MemStorage) GetGaugeValue(id string) float64 {
	return float64(s.gaugeData[id])
}

func (s *MemStorage) AllMetrics() string {
	var result string
	result += "Gauge metrics:\n"
	for n, v := range s.gaugeData {
		result += fmt.Sprintf("- %s = %f\n", n, v)
	}

	result += "Counter metrics:\n"
	for n, v := range s.counterData {
		result += fmt.Sprintf("- %s = %d\n", n, v)
	}

	return result
}

func (s *MemStorage) GetCounterData() map[string]counter {
	return s.counterData
}

func (s *MemStorage) GetGaugeData() map[string]gauge {
	return s.gaugeData
}

func (s *MemStorage) UpdateGaugeData(gaugeData map[string]gauge) {
	s.gaugeData = gaugeData
}

func (s *MemStorage) UpdateCounterData(counterData map[string]counter) {
	s.counterData = counterData
}
