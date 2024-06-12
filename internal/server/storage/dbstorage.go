package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zavtra-na-rabotu/gometrics/internal/model"
	"go.uber.org/zap"
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

func (storage *DBStorage) RunMigrations() error {
	driver, err := postgres.WithInstance(storage.DB, &postgres.Config{})
	if err != nil {
		zap.L().Fatal("Failed to create migration driver", zap.Error(err))
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance("file://internal/server/storage/migrations", "public", driver)
	if err != nil {
		zap.L().Fatal("Failed to create migrate instance", zap.Error(err))
		return err
	}
	err = migration.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		zap.L().Fatal("Failed to run migrations", zap.Error(err))
		return err
	}

	if errors.Is(err, migrate.ErrNoChange) {
		zap.L().Info("No migrations to run")
	} else {
		zap.L().Info("Successfully ran migrations")
	}

	return nil
}

func (storage *DBStorage) Close() {
	err := storage.DB.Close()
	if err != nil {
		zap.L().Fatal("Failed to close database", zap.Error(err))
	}
}

func (storage *DBStorage) Ping() error {
	return storage.DB.Ping()
}

func (storage *DBStorage) UpdateGauge(name string, metric float64) error {
	_, err := storage.DB.Exec(`
		INSERT INTO gauge (name, value) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;
	`, name, metric)
	if err != nil {
		zap.L().Error("Failed to update gauge metric", zap.Error(err))
		return err
	}

	return err
}

func (storage *DBStorage) UpdateCounter(name string, metric int64) error {
	_, err := storage.DB.Exec(`
		INSERT INTO counter (name, value) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = counter.value + EXCLUDED.value;
	`, name, metric)
	if err != nil {
		zap.L().Error("Failed to update counter metric", zap.Error(err))
		return err
	}

	return err
}

func (storage *DBStorage) UpdateCounterAndReturn(name string, metric int64) (int64, error) {
	err := storage.DB.QueryRow(`
		INSERT INTO counter (name, value) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = counter.value + EXCLUDED.value
		RETURNING value;
	`, name, metric).Scan(&metric)
	if err != nil {
		zap.L().Error("Failed to update and return counter metric", zap.Error(err))
		return 0, err
	}

	return metric, nil
}

func (storage *DBStorage) GetGauge(name string) (float64, error) {
	var value float64
	err := storage.DB.QueryRow(`SELECT value FROM gauge WHERE name = $1`, name).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrItemNotFound
		}
		zap.L().Error("Failed to select gauge metric", zap.Error(err))
		return 0, err
	}
	return value, nil
}

func (storage *DBStorage) GetCounter(name string) (int64, error) {
	var value int64
	err := storage.DB.QueryRow(`SELECT value FROM counter WHERE name = $1`, name).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrItemNotFound
		}
		zap.L().Error("Failed to select counter metric", zap.Error(err))
		return 0, err
	}
	return value, nil
}

func (storage *DBStorage) GetAllGauge() (map[string]float64, error) {
	rows, err := storage.DB.Query(`SELECT name, value FROM gauge`)
	if err != nil {
		zap.L().Error("Failed to get all gauge metrics")
		return nil, err
	}

	defer rows.Close()

	gaugeMetrics := make(map[string]float64)

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			zap.L().Error("Failed to get all gauge metrics", zap.Error(err))
			return nil, err
		}
		gaugeMetrics[name] = value
	}

	if err := rows.Err(); err != nil {
		zap.L().Error("Failed to get all gauge metrics", zap.Error(err))
		return nil, err
	}

	return gaugeMetrics, nil
}

func (storage *DBStorage) GetAllCounter() (map[string]int64, error) {
	rows, err := storage.DB.Query(`SELECT name, value FROM counter`)
	if err != nil {
		zap.L().Error("Failed to get all counter metrics")
		return nil, err
	}

	defer rows.Close()

	counterMetrics := make(map[string]int64)

	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			zap.L().Error("Failed to get all counter metrics", zap.Error(err))
			return nil, err
		}
		counterMetrics[name] = value
	}

	if err := rows.Err(); err != nil {
		zap.L().Error("Failed to get all counter metrics", zap.Error(err))
		return nil, err
	}

	return counterMetrics, nil
}

func (storage *DBStorage) UpdateMetrics(metrics []model.Metrics) error {
	tx, err := storage.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				zap.L().Error("Failed to rollback transaction", zap.Error(err))
			}
		} else {
			err := tx.Commit()
			if err != nil {
				zap.L().Error("Failed to commit transaction", zap.Error(err))
			}
		}
	}()

	for _, metric := range metrics {
		switch metric.MType {
		case string(model.Gauge):
			err := updateGaugeInTransaction(tx, metric.ID, *metric.Value)
			if err != nil {
				return err
			}
		case string(model.Counter):
			err := updateCounterInTransaction(tx, metric.ID, *metric.Delta)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown metric type: %s", metric.MType)
		}
	}

	return nil
}

/*
TODO: Я бы очень хотел написать враппер над уже существующими методами выше, но не понимаю как
TODO: Если первым аргументом я везде добавлю Tx, то нарушу сигнатуру метода основного интерфейса Storage
*/
func updateGaugeInTransaction(tx *sql.Tx, name string, metric float64) error {
	_, err := tx.Exec(`
		INSERT INTO gauge (name, value) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;
	`, name, metric)
	if err != nil {
		zap.L().Error("Failed to update gauge metric in transaction", zap.Error(err))
		return err
	}

	return err
}

func updateCounterInTransaction(tx *sql.Tx, name string, metric int64) error {
	_, err := tx.Exec(`
		INSERT INTO counter (name, value) VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = counter.value + EXCLUDED.value;
	`, name, metric)
	if err != nil {
		zap.L().Error("Failed to update counter metric in transaction", zap.Error(err))
		return err
	}

	return err
}
