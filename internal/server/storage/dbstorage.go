package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	DB *sql.DB
}

type Repository interface {
	Ping() error
}

func NewDBStorage(databaseDsn string) (*DBStorage, error) {
	db, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		return nil, err
	}
	return &DBStorage{DB: db}, nil
}

func (s *DBStorage) Ping() error {
	return s.DB.Ping()
}
