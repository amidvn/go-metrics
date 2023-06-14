package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/amidvn/go-metrics/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
)

type DBConnection struct {
	DB *sql.DB
}

type counterMetric struct {
	name  string
	value int64
}

type gaugeMetric struct {
	name  string
	value float64
}

func New(dsn string) *DBConnection {
	dbc := &DBConnection{}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fmt.Println(err)
		dbc.DB = nil
	} else {
		dbc.DB = db
	}

	// checkint if tables exist or not
	_, table_check := dbc.DB.Query("SELECT * FROM counter_metrics LIMIT 1;")
	if table_check != nil {
		dbc.DB.Exec("CREATE TABLE counter_metrics (name char(30) UNIQUE, value integer);")
	}
	_, table_check = dbc.DB.Query("SELECT * FROM gauge_metrics LIMIT 1;")
	if table_check != nil {
		dbc.DB.Exec("CREATE TABLE gauge_metrics (name char(30) UNIQUE, value double precision);")
	}
	return dbc
}

func CheckConnection(dbc *DBConnection) error {
	if dbc.DB != nil {
		err := dbc.DB.Ping()
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Empty connection string")
}

func Restore(s *storage.MemStorage, dbc *DBConnection) {
	rowsCounter, err := dbc.DB.Query("SELECT name, value FROM counter_metrics;")
	if err != nil {
		fmt.Println(err)
	}
	defer rowsCounter.Close()

	for rowsCounter.Next() {
		var cm counterMetric
		err = rowsCounter.Scan(&cm.name, &cm.value)
		if err != nil {
			fmt.Println(err)
		}
		s.UpdateCounter(cm.name, cm.value)
	}

	rowsGauge, err := dbc.DB.Query("SELECT name, value FROM gauge_metrics;")
	if err != nil {
		fmt.Println(err)
	}
	defer rowsGauge.Close()

	for rowsGauge.Next() {
		var gm gaugeMetric
		err = rowsGauge.Scan(&gm.name, &gm.value)
		if err != nil {
			fmt.Println(err)
		}
		s.UpdateGauge(gm.name, gm.value)
	}
}

func Dump(s *storage.MemStorage, dbc *DBConnection, storeInterval int) {
	pollTicker := time.NewTicker(time.Duration(storeInterval) * time.Second)
	defer pollTicker.Stop()
	for range pollTicker.C {
		saveMetrics(s, dbc)
	}
}

func saveMetrics(s *storage.MemStorage, dbc *DBConnection) error {
	var query string
	query = "TRUNCATE counter_metrics, gauge_metrics; "
	for k, v := range s.GetCounterData() {
		query += fmt.Sprintf("INSERT INTO counter_metrics (name, value) VALUES ('%s', %d); ", k, v)
	}

	for k, v := range s.GetGaugeData() {
		query += fmt.Sprintf("INSERT INTO gauge_metrics (name, value) VALUES ('%s', %f); ", k, v)
	}

	_, err := dbc.DB.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
