package filestoring

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/amidvn/go-metrics/internal/storage"
)

func Restore(s *storage.MemStorage, filePath string) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}

	var data storage.AllMetrics
	if err := json.Unmarshal(file, &data); err != nil {
		fmt.Println(err)
	}

	if len(data.Counter) != 0 {
		s.UpdateCounterData(data.Counter)
	}
	if len(data.Gauge) != 0 {
		s.UpdateGaugeData(data.Gauge)
	}
}

func Dump(s *storage.MemStorage, filePath string, storeInterval int) {
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
		saveJSON(s, filePath)
	}
}

func saveJSON(s *storage.MemStorage, filePath string) error {

	var metrics storage.AllMetrics

	metrics.Counter = s.GetCounterData()
	metrics.Gauge = s.GetGaugeData()

	data, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0666)
}
