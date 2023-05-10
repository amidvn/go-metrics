package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/levigross/grequests"
)

var pollInterval int
var reportInterval int
var addressServer string

var metrics = map[string]bool{
	"Alloc":         true,
	"BuckHashSys":   true,
	"Frees":         true,
	"GCCPUFraction": true,
	"GCSys":         true,
	"HeapAlloc":     true,
	"HeapIdle":      true,
	"HeapInuse":     true,
	"HeapObjects":   true,
	"HeapReleased":  true,
	"HeapSys":       true,
	"LastGC":        true,
	"Lookups":       true,
	"MCacheInuse":   true,
	"MCacheSys":     true,
	"MSpanInuse":    true,
	"MSpanSys":      true,
	"Mallocs":       true,
	"NextGC":        true,
	"NumForcedGC":   true,
	"NumGC":         true,
	"OtherSys":      true,
	"PauseTotalNs":  true,
	"StackInuse":    true,
	"StackSys":      true,
	"Sys":           true,
	"TotalAlloc":    true,
	"PollCount":     true,
	"RandomValue":   true,
}

var valuesGauge = map[string]float64{}
var pollCount uint64

func main() {
	getParameters()

	pollTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			fmt.Println("get metrics ", pollCount)
			getMetrics()
		case <-reportTicker.C:
			fmt.Println("send post")
			postQueries()
		}
	}
}

func getParameters() {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addressServer = envRunAddr
	} else {
		flag.StringVar(&addressServer, "a", "localhost:8080", "address and port to run server")
	}

	if envRunAddr := os.Getenv("REPORT_INTERVAL"); envRunAddr != "" {
		reportInterval, _ = strconv.Atoi(envRunAddr)
	} else {
		flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	}

	if envRunAddr := os.Getenv("POLL_INTERVAL"); envRunAddr != "" {
		pollInterval, _ = strconv.Atoi(envRunAddr)
	} else {
		flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	}
	flag.Parse()
}

func getMetrics() {
	var rtm runtime.MemStats

	pollCount += 1
	runtime.ReadMemStats(&rtm)
	numfield := reflect.ValueOf(&rtm).Elem().NumField()
	for x := 0; x < numfield; x++ {
		metricsName := reflect.TypeOf(&rtm).Elem().Field(x).Name
		if metrics[metricsName] {
			metricsValue := reflect.ValueOf(&rtm).Elem().Field(x)
			var metricsFloat float64
			if metricsValue.CanFloat() {
				metricsFloat = float64(metricsValue.Float())
			} else if metricsValue.CanUint() {
				metricsFloat = float64(metricsValue.Uint())
			}
			valuesGauge[metricsName] = metricsFloat
		}
	}
}

func postQueries() {
	for k, v := range valuesGauge {
		post("gauge", k, strconv.FormatFloat(v, 'f', -1, 64))
	}
	post("counter", "PollCount", strconv.FormatUint(pollCount, 10))
	post("gauge", "RandomValue", strconv.FormatFloat(rand.Float64(), 'f', -1, 64))
	pollCount = 0
}

func post(t string, mn string, sValue string) {
	_, err := grequests.Post(fmt.Sprintf("http://%s/update/%s/%s/%s", addressServer, t, mn, sValue),
		&grequests.RequestOptions{
			Headers: map[string]string{"content-type": "text/plain"},
		})
	if err != nil {
		log.Fatal(err)
	}
}
