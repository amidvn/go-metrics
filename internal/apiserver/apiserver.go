package apiserver

import (
	"flag"
	"fmt"
	"log"

	"github.com/amidvn/go-metrics/internal/database"
	"github.com/amidvn/go-metrics/internal/handlers"
	"github.com/amidvn/go-metrics/internal/middlewares"
	"github.com/amidvn/go-metrics/internal/storage"
	"github.com/amidvn/go-metrics/internal/storing"
	"github.com/caarlos0/env/v6"
	"github.com/labstack/echo/v4"

	"go.uber.org/zap"
)

type Conf struct {
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	DbDSN         string `env:"DATABASE_DSN"`
}

type APIServer struct {
	storage *storage.MemStorage
	echo    *echo.Echo
	address string
	logger  zap.SugaredLogger
	config  *Conf
	db      *database.DBConnection
}

func New() *APIServer {
	a := &APIServer{}

	var conf Conf
	flag.StringVar(&conf.Address, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&conf.StoreInterval, "i", 300, "interval for saving metrics on the server")
	flag.StringVar(&conf.FilePath, "f", "/tmp/metrics-db.json", "file storage path for saving data")
	flag.BoolVar(&conf.Restore, "r", true, "need to load data at startup")
	flag.StringVar(&conf.DbDSN, "d", "", "file storage path for saving data")
	flag.Parse()

	err := env.Parse(&conf)
	if err != nil {
		fmt.Println(err)
	}

	a.address = conf.Address
	a.config = &conf

	a.storage = storage.New(conf.StoreInterval, conf.FilePath, conf.Restore)
	a.echo = echo.New()

	a.db = database.New(conf.DbDSN)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	a.logger = *logger.Sugar()

	if conf.FilePath != "" {
		if conf.Restore {
			storing.Restore(a.storage, conf.FilePath)
		}
		if conf.StoreInterval != 0 {
			go storing.Dump(a.storage, conf.FilePath, conf.StoreInterval)
		}
	}

	a.echo.Use(middlewares.WithLogging(a.logger))
	a.echo.Use(middlewares.GzipUnpacking())

	a.echo.GET("/", handlers.AllMetrics(a.storage))
	a.echo.POST("/value/", handlers.GetValueJSON(a.storage))
	a.echo.GET("/value/:typeM/:nameM", handlers.MetricsValue(a.storage))
	a.echo.POST("/update/", handlers.UpdateJSON(a.storage))
	a.echo.POST("/update/:typeM/:nameM/:valueM", handlers.PostWebhook(a.storage))
	a.echo.GET("/ping", handlers.PingDB(a.db))

	return a
}

func (a *APIServer) Start() error {
	err := a.echo.Start(a.address)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
