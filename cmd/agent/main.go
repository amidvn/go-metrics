package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/amidvn/go-metrics/internal/models"
	"github.com/levigross/grequests"
)

type Config struct {
	pollInterval   int    `env:"POLL_INTERVAL"`
	reportInterval int    `env:"REPORT_INTERVAL"`
	addressServer  string `env:"ADDRESS"`
}

var cfg Config

var valuesGauge = map[string]float64{}
var pollCount uint64

func main() {
	err := getParameters()
	if err != nil {
		log.Fatal(err)
	}

	pollTicker := time.NewTicker(time.Duration(cfg.pollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(cfg.reportInterval) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			getMetrics()
		case <-reportTicker.C:
			postQueries()
		}
	}
}

func getParameters() error {
	flag.StringVar(&cfg.addressServer, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.reportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&cfg.pollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.addressServer = envRunAddr
	}
	if envRunAddr := os.Getenv("REPORT_INTERVAL"); envRunAddr != "" {
		cfg.reportInterval, _ = strconv.Atoi(envRunAddr)
	}
	if envRunAddr := os.Getenv("POLL_INTERVAL"); envRunAddr != "" {
		cfg.pollInterval, _ = strconv.Atoi(envRunAddr)
	}
	return nil
}

func getMetrics() {
	var rtm runtime.MemStats

	pollCount += 1
	runtime.ReadMemStats(&rtm)

	valuesGauge["Alloc"] = float64(rtm.Alloc)
	valuesGauge["BuckHashSys"] = float64(rtm.BuckHashSys)
	valuesGauge["Frees"] = float64(rtm.Frees)
	valuesGauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	valuesGauge["HeapAlloc"] = float64(rtm.HeapAlloc)
	valuesGauge["HeapIdle"] = float64(rtm.HeapIdle)
	valuesGauge["HeapInuse"] = float64(rtm.HeapInuse)
	valuesGauge["HeapObjects"] = float64(rtm.HeapObjects)
	valuesGauge["HeapReleased"] = float64(rtm.HeapReleased)
	valuesGauge["HeapSys"] = float64(rtm.HeapSys)
	valuesGauge["LastGC"] = float64(rtm.LastGC)
	valuesGauge["Lookups"] = float64(rtm.Lookups)
	valuesGauge["MCacheInuse"] = float64(rtm.MCacheInuse)
	valuesGauge["MCacheSys"] = float64(rtm.MCacheSys)
	valuesGauge["MSpanInuse"] = float64(rtm.MSpanInuse)
	valuesGauge["MSpanSys"] = float64(rtm.MSpanSys)
	valuesGauge["Mallocs"] = float64(rtm.Mallocs)
	valuesGauge["NextGC"] = float64(rtm.NextGC)
	valuesGauge["NumForcedGC"] = float64(rtm.NumForcedGC)
	valuesGauge["NumGC"] = float64(rtm.NumGC)
	valuesGauge["OtherSys"] = float64(rtm.OtherSys)
	valuesGauge["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	valuesGauge["StackInuse"] = float64(rtm.StackInuse)
	valuesGauge["StackSys"] = float64(rtm.StackSys)
	valuesGauge["Sys"] = float64(rtm.Sys)
	valuesGauge["TotalAlloc"] = float64(rtm.TotalAlloc)
}

func postQueries() {
	url := fmt.Sprintf("http://%s/update/", cfg.addressServer)
	ro := grequests.RequestOptions{
		Headers: map[string]string{
			"content-type":     "application/json",
			"content-encoding": "gzip",
		},
	}
	session := grequests.NewSession(&ro)
	for k, v := range valuesGauge {
		postJSON(session, url, models.Metrics{ID: k, MType: "gauge", Value: &v})
	}
	pc := int64(pollCount)
	postJSON(session, url, models.Metrics{ID: "PollCount", MType: "counter", Delta: &pc})
	r := rand.Float64()
	postJSON(session, url, models.Metrics{ID: "RandomValue", MType: "gauge", Value: &r})
	pollCount = 0
}

func postJSON(s *grequests.Session, url string, m models.Metrics) {
	js, err := json.Marshal(&m)
	if err != nil {
		fmt.Println(err)
	}

	gz, err := compress(js)
	if err != nil {
		fmt.Println(err)
	}

	resp, err := s.Post(url, &grequests.RequestOptions{JSON: gz})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.StatusCode)
}

func compress(b []byte) ([]byte, error) {
	var bf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&bf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	_, err = gz.Write(b)
	if err != nil {
		return nil, err
	}
	gz.Close()
	return bf.Bytes(), nil
}
