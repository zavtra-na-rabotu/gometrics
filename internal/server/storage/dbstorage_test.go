package storage

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDBStorage_Ping(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectPing().WillReturnError(nil)

	storage := &DBStorage{DB: db}
	err = storage.Ping()
	assert.NoError(t, err)
}

func TestDBStorage_UpdateGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta(`INSERT INTO gauge (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;`)
	mock.ExpectExec(query).WithArgs("test_metric", 123.45).
		WillReturnResult(sqlmock.NewResult(1, 1))

	storage := &DBStorage{DB: db}
	err = storage.UpdateGauge("test_metric", 123.45)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_UpdateCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta(`INSERT INTO counter (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = counter.value + EXCLUDED.value;`)
	mock.ExpectExec(query).WithArgs("test_metric", int64(10)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	storage := &DBStorage{DB: db}
	err = storage.UpdateCounter("test_metric", 10)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_UpdateCounterAndReturn(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	expectedValue := int64(10)
	query := regexp.QuoteMeta(`INSERT INTO counter (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value = counter.value + EXCLUDED.value RETURNING value;`)
	mock.ExpectQuery(query).WithArgs("test_metric", int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(expectedValue))

	storage := &DBStorage{DB: db}
	value, err := storage.UpdateCounterAndReturn("test_metric", 10)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_GetGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	expectedValue := 123.45
	rows := sqlmock.NewRows([]string{"value"}).AddRow(expectedValue)
	query := regexp.QuoteMeta(`SELECT value FROM gauge WHERE name = $1`)
	mock.ExpectQuery(query).WithArgs("test_metric").WillReturnRows(rows)

	storage := &DBStorage{DB: db}
	value, err := storage.GetGauge("test_metric")
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_GetCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	expectedValue := int64(123)
	rows := sqlmock.NewRows([]string{"value"}).AddRow(expectedValue)
	query := regexp.QuoteMeta(`SELECT value FROM counter WHERE name = $1`)
	mock.ExpectQuery(query).WithArgs("test_metric").WillReturnRows(rows)

	storage := &DBStorage{DB: db}
	value, err := storage.GetCounter("test_metric")
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_GetAllGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("metric1", 100.0).
		AddRow("metric2", 200.0)
	query := regexp.QuoteMeta(`SELECT name, value FROM gauge`)
	mock.ExpectQuery(query).WillReturnRows(rows)

	storage := &DBStorage{DB: db}
	gauges, err := storage.GetAllGauge()
	assert.NoError(t, err)
	assert.Equal(t, map[string]float64{"metric1": 100.0, "metric2": 200.0}, gauges)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestDBStorage_GetAllCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("counter1", 10).
		AddRow("counter2", 20)
	query := regexp.QuoteMeta(`SELECT name, value FROM counter`)
	mock.ExpectQuery(query).WillReturnRows(rows)

	storage := DBStorage{DB: db}
	counters, err := storage.GetAllCounter()
	assert.NoError(t, err)
	assert.Equal(t, map[string]int64{"counter1": 10, "counter2": 20}, counters)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
