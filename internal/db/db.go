package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DBConnector struct {
	DSN        string
	Ctx        context.Context
	Pool       *pgxpool.Pool
	initalized bool
}

func (d *DBConnector) checkInit() error {
	if !d.initalized {
		err := fmt.Errorf("DbConnector is not initiliazed (run DBConnector.Init() to initilize)")
		return err
	}
	return nil
}

func (d *DBConnector) Init() error {
	p, err := pgxpool.Connect(d.Ctx, d.DSN)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	d.Pool = p
	err = d.CreateTables()
	if err != nil {
		return err
	}
	d.initalized = true
	return nil
}

func (d *DBConnector) Close() {
	if d.initalized {
		d.Pool.Close()
		d.initalized = false
	}
}

func (d *DBConnector) Avaliable() error {
	err := d.checkInit()
	if err != nil {
		return err
	}
	return d.Pool.Ping(d.Ctx)
}

func (d *DBConnector) GetGauge(metricId string) (float64, error) {
	err := d.checkInit()
	if err != nil {
		return -1, err
	}
	var gauge float64
	sql := `SELECT metric_value FROM gauges WHERE metric_id=$1;`
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	row := conn.QueryRow(d.Ctx, sql, metricId)
	switch err := row.Scan(&gauge); err {
	case pgx.ErrNoRows:
		return -1, fmt.Errorf("no metric %s was found", metricId)
	case nil:
		return gauge, nil
	default:
		e := fmt.Errorf("unknown error while quering metric %s: %s", metricId, err.Error())
		return -1, e
	}
}

func (d *DBConnector) SetGauge(metricId string, metricValue float64) error {
	err := d.checkInit()
	if err != nil {
		return err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	sql := `INSERT INTO gauges (metric_id, metric_value)
			VALUES($1, $2) 
			ON CONFLICT (metric_id) 
			DO 
				UPDATE SET metric_value = $2 WHERE gauges.metric_id = $1;`
	_, err = conn.Exec(d.Ctx, sql, metricId, metricValue)
	return err
}

func (d *DBConnector) GetAllGauges() (map[string]float64, error) {
	err := d.checkInit()
	if err != nil {
		return make(map[string]float64), err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return make(map[string]float64), fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	gauges := make(map[string]float64)
	sql := "SELECT metric_id, metric_value FROM gauges"
	rows, err := conn.Query(d.Ctx, sql)
	if err != nil {
		e := fmt.Errorf("failed to query gauges table: %s", err.Error())
		return make(map[string]float64), e
	}
	defer rows.Close()

	for rows.Next() {
		var metricId string
		var metricValue float64
		if err := rows.Scan(&metricId, &metricValue); err != nil {
			e := fmt.Errorf("failed to convert row to map entry: %s", err.Error())
			return make(map[string]float64), e
		}
		gauges[metricId] = metricValue
	}

	if err := rows.Err(); err != nil {
		e := fmt.Errorf("error(s) occured during gauges table scanning: %s", err.Error())
		return make(map[string]float64), e
	}

	return gauges, nil
}

func (d *DBConnector) GetCounter(metricId string) (int64, error) {
	err := d.checkInit()
	if err != nil {
		return -1, err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	var counter int64
	sql := `SELECT metric_value FROM counters WHERE metric_id=$1;`
	row := conn.QueryRow(d.Ctx, sql, metricId)
	switch err := row.Scan(&counter); err {
	case pgx.ErrNoRows:
		e := fmt.Errorf("no metric %s was found", metricId)
		return -1, e
	case nil:
		return counter, nil
	default:
		e := fmt.Errorf("unknown error while quering metric %s: %s", metricId, err.Error())
		return -1, e
	}
}

func (d *DBConnector) SetCounter(metricId string, metricValue int64) error {
	err := d.checkInit()
	if err != nil {
		return err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	sql := `INSERT INTO counters (metric_id, metric_value)
			VALUES($1, $2) 
			ON CONFLICT (metric_id) 
			DO 
				UPDATE SET metric_value = $2 WHERE counters.metric_id = $1;`
	_, err = conn.Exec(d.Ctx, sql, metricId, metricValue)
	return err
}
func (d *DBConnector) IncreaseCounter(metricId string, metricValue int64) error {
	err := d.checkInit()
	if err != nil {
		return err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	sql := `INSERT INTO counters (metric_id, metric_value)
			VALUES($1, $2) 
			ON CONFLICT (metric_id)
			DO
				UPDATE SET metric_value = counters.metric_value + $2
				WHERE counters.metric_id = $1;`
	_, err = conn.Query(d.Ctx, sql, metricId, metricValue)
	return err
}
func (d *DBConnector) GetAllCounters() (map[string]int64, error) {
	err := d.checkInit()
	if err != nil {
		return make(map[string]int64), err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return make(map[string]int64), fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	counters := make(map[string]int64)
	sql := "SELECT metric_id, metric_value FROM counters"
	rows, err := conn.Query(d.Ctx, sql)
	if err != nil {
		e := fmt.Errorf("failed to query counters table: %s", err.Error())
		return make(map[string]int64), e
	}
	defer rows.Close()

	for rows.Next() {
		var metricId string
		var metricValue int64
		if err := rows.Scan(&metricId, &metricValue); err != nil {
			e := fmt.Errorf("failed to convert row to map entry: %s", err.Error())
			return make(map[string]int64), e
		}
		counters[metricId] = metricValue
	}

	if err := rows.Err(); err != nil {
		e := fmt.Errorf("error(s) occured during counters table scanning: %s", err.Error())
		return make(map[string]int64), e
	}

	return counters, nil
}

func (d *DBConnector) ResetCounter(metricId string) error {
	err := d.checkInit()
	if err != nil {
		return err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()
	sql := `UPDATE counters 
			SET metric_value = 0
			WHERE counters.metric_id = $1;`
	_, err = conn.Exec(d.Ctx, sql, metricId)
	return err
}

func (d *DBConnector) CreateTables() error {
	conn, err := d.Pool.Acquire(d.Ctx)
	defer conn.Release()
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	countersSQL := `CREATE TABLE IF NOT EXISTS counters(
		metric_id varchar(45) NOT NULL,
		metric_value bigint NOT NULL,
		PRIMARY KEY (metric_id)
	  )`

	gaugesSQL := `CREATE TABLE IF NOT EXISTS gauges(
		metric_id varchar(45) NOT NULL,
		metric_value double precision NOT NULL,
		PRIMARY KEY (metric_id)
	)`

	_, err = conn.Exec(d.Ctx, countersSQL)
	if err != nil {
		return fmt.Errorf("cant create counters table: %s", err.Error())
	}

	_, err = conn.Exec(d.Ctx, gaugesSQL)
	if err != nil {
		return fmt.Errorf("cant create пфгпуы table: %s", err.Error())
	}
	return nil
}
