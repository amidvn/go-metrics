package apiserver

import (
	"flag"
	"log"
	"os"

	"github.com/amidvn/go-metrics/internal/handlers"
	"github.com/amidvn/go-metrics/internal/storage"
	"github.com/labstack/echo/v4"
)

type APIServer struct {
	storage *storage.MemStorage
	echo    *echo.Echo
	address string
}

func New() *APIServer {
	a := &APIServer{}
	a.storage = storage.New()
	a.echo = echo.New()

	var address string
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		address = envRunAddr
	} else {
		flag.StringVar(&address, "a", "localhost:8080", "address and port to run server")
		flag.Parse()
	}
	a.address = address

	a.echo.GET("/", handlers.AllMetrics(a.storage))
	a.echo.GET("/value/:typeM/:nameM", handlers.MetricsValue(a.storage))
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
