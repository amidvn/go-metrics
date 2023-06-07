package apiserver

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/amidvn/go-metrics/internal/storage"
)

func loadStorageFromFile(s *storage.MemStorage, filePath string) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}

	var data storage.AllMetrics
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Println(err)
	}

	s.UpdateCounterData(data.Counter)
	s.UpdateGaugeData(data.Gauge)
}

func storing(s *storage.MemStorage, filePath string, storeInterval int) {
	dir, _ := path.Split(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0666)
		if err != nil {
			fmt.Println(err)
		}
	}
	pollTicker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	defer pollTicker.Stop()
	for range pollTicker.C {
		saveStorageToFile(s, filePath)
	}
}

func saveStorageToFile(s *storage.MemStorage, filePath string) error {

	var metrics storage.AllMetrics

	metrics.Counter = s.GetCounterData()
	metrics.Gauge = s.GetGaugeData()

	data, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0666)
}
