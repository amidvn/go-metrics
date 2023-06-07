package apiserver

import (
	"flag"
	"fmt"
	"log"

	"github.com/amidvn/go-metrics/internal/handlers"
	"github.com/amidvn/go-metrics/internal/middlewares"
	"github.com/amidvn/go-metrics/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/labstack/echo/v4"

	"go.uber.org/zap"
)

type Conf struct {
	address       string
	storeInterval int
	filePath      string
	restore       bool
}

type APIServer struct {
	storage *storage.MemStorage
	echo    *echo.Echo
	address string
	logger  zap.SugaredLogger
	config  *Conf
}

func New() *APIServer {
	a := &APIServer{}

	var conf Conf
	flag.StringVar(&conf.address, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&conf.storeInterval, "i", 300, "interval for saving metrics on the server")
	flag.StringVar(&conf.filePath, "f", "/tmp/metrics-db.json", "file storage path for saving data")
	flag.BoolVar(&conf.restore, "r", true, "need to load data at startup")
	flag.Parse()

	err := env.Parse(&conf)
	if err != nil {
		fmt.Println(err)
	}

	a.address = conf.address
	a.config = &conf

	a.storage = storage.New(conf.storeInterval, conf.filePath, conf.restore)
	a.echo = echo.New()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	a.logger = *logger.Sugar()

	if conf.filePath != "" {
		if conf.restore {
			loadStorageFromFile(a.storage, conf.filePath)
		}
		if conf.storeInterval != 0 {
			go storing(a.storage, conf.filePath, conf.storeInterval)
		}
	}

	a.echo.Use(middlewares.WithLogging(a.logger))
	a.echo.Use(middlewares.GzipUnpacking())

	a.echo.GET("/", handlers.AllMetrics(a.storage))
	a.echo.POST("/value/", handlers.GetValueJSON(a.storage))
	a.echo.GET("/value/:typeM/:nameM", handlers.MetricsValue(a.storage))
	a.echo.POST("/update/", handlers.UpdateJSON(a.storage))
	a.echo.POST("/update/:typeM/:nameM/:valueM", handlers.PostWebhook(a.storage))

	return a
}

func (a *APIServer) Start() error {
	err := a.echo.Start(a.address)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
