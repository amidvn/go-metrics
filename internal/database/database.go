package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

type DBConnection struct {
	DSN string
}

func New(dsn string) *DBConnection {
	db := &DBConnection{}
	db.DSN = dsn

	return db
}

func CheckConnection(db *DBConnection) error {
	if db.DSN == "" {
		return errors.New("Empty connection string")
	}
	dbc, err := sql.Open("pgx", db.DSN)
	if err != nil {
		return err
	}
	defer dbc.Close()
	return nil
}
