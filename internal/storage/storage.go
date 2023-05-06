package storage

type gauge float64
type counter int64

type MemStorage struct {
	gaugeData   map[string]gauge
	counterData map[string]counter
}

func New() *MemStorage {
	return &MemStorage{
		gaugeData:   make(map[string]gauge),
		counterData: make(map[string]counter),
	}
}

func (s *MemStorage) UpdateCounter(n string, v int64) {
	s.counterData[n] += counter(v)
}

func (s *MemStorage) UpdateGauge(n string, v float64) {
	s.gaugeData[n] = gauge(v)
}
