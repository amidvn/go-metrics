package storage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"
)

type gauge float64
type counter int64

type MemStorage struct {
	gaugeData     map[string]gauge
	counterData   map[string]counter
	storeInterval int
	filePath      string
}

type allMetrics struct {
	Gauge   map[string]gauge   `json:"gauge"`
	Counter map[string]counter `json:"counter"`
}

func New(storeInterval int, filePath string, restore bool) *MemStorage {
	storage := MemStorage{
		gaugeData:     make(map[string]gauge),
		counterData:   make(map[string]counter),
		storeInterval: storeInterval,
		filePath:      filePath,
	}

	if filePath != "" {
		if restore {
			storage.loadFromFile(filePath)
		}
		if storeInterval != 0 {
			go storage.storing()
		}
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

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) (*Consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (c *Consumer) readMetrics(s *MemStorage) {
	var data allMetrics
	err := c.decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.counterData = data.Counter
	s.gaugeData = data.Gauge
}

func (s *MemStorage) loadFromFile(filePath string) {
	Consumer, err := NewConsumer(filePath)
	if err != nil {
		fmt.Println(err)
	}
	defer Consumer.Close()

	Consumer.readMetrics(s)
}

func (s *MemStorage) storing() {
	dir, _ := path.Split(s.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0666)
		if err != nil {
			fmt.Println(err)
		}
	}
	pollTicker := time.NewTicker(time.Duration(s.storeInterval) * time.Second)
	defer pollTicker.Stop()
	for range pollTicker.C {
		s.saveMetrics()
	}
}

func (s *MemStorage) saveMetrics() error {
	if err := s.saveToFile(); err != nil {
		fmt.Printf("Error in storing to file: %v", err)
		return err
	}
	return nil
}

func (s *MemStorage) saveToFile() error {
	path := s.filePath
	p, err := NewProducer(path)
	if err != nil {
		return err
	}
	defer p.Close()
	var metrics allMetrics

	metrics.Counter = s.counterData
	metrics.Gauge = s.gaugeData

	return p.encoder.Encode(&metrics)
}
