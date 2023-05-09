package apiserver

import (
	"log"

	"github.com/amidvn/go-metrics/internal/handlers"
	"github.com/amidvn/go-metrics/internal/storage"
	"github.com/labstack/echo/v4"
)

type APIServer struct {
	storage *storage.MemStorage
	echo    *echo.Echo
}

func New() *APIServer {
	a := &APIServer{}
	a.storage = storage.New()
	a.echo = echo.New()

	a.echo.GET("/", handlers.AllMetrics(a.storage))
	a.echo.GET("/value/:typeM/:nameM", handlers.MetricsValue(a.storage))
	a.echo.POST("/update/:typeM/:nameM/:valueM", handlers.PostWebhook(a.storage))

	return a
}

func (a *APIServer) Start() error {
	err := a.echo.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
