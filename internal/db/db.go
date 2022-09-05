// Package db responsible for working with Postgresql database
package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zklevsha/go-musthave-devops/internal/structs"
)

type DBConnector struct {
	Ctx         context.Context
	Pool        *pgxpool.Pool
	DSN         string
	initialized bool
}

func (d *DBConnector) checkInit() error {
	if !d.initialized {
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
	d.initialized = true
	return nil
}

func (d *DBConnector) Close() {
	if d.initialized {
		d.Pool.Close()
		d.initialized = false
	}
}

func (d *DBConnector) Avaliable() error {
	err := d.checkInit()
	if err != nil {
		return err
	}
	return d.Pool.Ping(d.Ctx)
}

func (d *DBConnector) getGauge(metricID string) (float64, error) {
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
	row := conn.QueryRow(d.Ctx, sql, metricID)
	switch err := row.Scan(&gauge); err {
	case pgx.ErrNoRows:
		return -1, structs.ErrMetricNotFound
	case nil:
		return gauge, nil
	default:
		e := fmt.Errorf("unknown error while quering metric %s: %s", metricID, err.Error())
		return -1, e
	}
}

func (d *DBConnector) setGauge(metricID string, metricValue float64) error {
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
	_, err = conn.Exec(d.Ctx, sql, metricID, metricValue)
	return err
}

func (d *DBConnector) getAllGauges() (map[string]float64, error) {
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
		var metricID string
		var metricValue float64
		if err := rows.Scan(&metricID, &metricValue); err != nil {
			e := fmt.Errorf("failed to convert row to map entry: %s", err.Error())
			return make(map[string]float64), e
		}
		gauges[metricID] = metricValue
	}

	if err := rows.Err(); err != nil {
		e := fmt.Errorf("error(s) occured during gauges table scanning: %s", err.Error())
		return make(map[string]float64), e
	}

	return gauges, nil
}

func (d *DBConnector) getCounter(metricID string) (int64, error) {
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
	row := conn.QueryRow(d.Ctx, sql, metricID)
	switch err := row.Scan(&counter); err {
	case pgx.ErrNoRows:
		return -1, structs.ErrMetricNotFound
	case nil:
		return counter, nil
	default:
		e := fmt.Errorf("unknown error while quering metric %s: %s", metricID, err.Error())
		return -1, e
	}
}

func (d *DBConnector) increaseCounter(metricID string, metricValue int64) error {
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
	_, err = conn.Query(d.Ctx, sql, metricID, metricValue)
	return err
}
func (d *DBConnector) getAllCounters() (map[string]int64, error) {
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
		var metricID string
		var metricValue int64
		if err := rows.Scan(&metricID, &metricValue); err != nil {
			e := fmt.Errorf("failed to convert row to map entry: %s", err.Error())
			return make(map[string]int64), e
		}
		counters[metricID] = metricValue
	}

	if err := rows.Err(); err != nil {
		e := fmt.Errorf("error(s) occured during counters table scanning: %s", err.Error())
		return make(map[string]int64), e
	}

	return counters, nil
}

func (d *DBConnector) ResetCounter(metricID string) error {
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
	_, err = conn.Exec(d.Ctx, sql, metricID)
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
		return fmt.Errorf("cant create gauge table: %s", err.Error())
	}
	return nil
}

func (d *DBConnector) DropTables() error {
	conn, err := d.Pool.Acquire(d.Ctx)
	defer conn.Release()
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}

	countersSQL := "DROP TABLE IF EXISTS counters"
	gaugesSQL := "DROP TABLE IF EXISTS gauges"

	_, err = conn.Exec(d.Ctx, countersSQL)
	if err != nil {
		return fmt.Errorf("cant drop counters table: %s", err.Error())
	}

	_, err = conn.Exec(d.Ctx, gaugesSQL)
	if err != nil {
		return fmt.Errorf("cant drop gauge table: %s", err.Error())
	}
	return nil
}

func (d *DBConnector) UpdateMetrics(metrics []structs.Metric) error {
	var err error
	err = d.checkInit()
	if err != nil {
		return err
	}
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	sqlCounters := `INSERT INTO counters (metric_id, metric_value)
					VALUES($1, $2) 
					ON CONFLICT (metric_id)
					DO
						UPDATE SET metric_value = counters.metric_value + $2
						WHERE counters.metric_id = $1;`
	sqlGauges := `INSERT INTO gauges (metric_id, metric_value)
				  VALUES($1, $2) 
				  ON CONFLICT (metric_id) 
				  DO 
					UPDATE SET metric_value = $2 WHERE gauges.metric_id = $1;`
	tx, err := conn.Begin(d.Ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %s", err.Error())
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(d.Ctx)

	for _, m := range metrics {
		switch m.MType {
		case "counter":
			_, err = tx.Exec(d.Ctx, sqlCounters, m.ID, m.Delta)
			if err != nil {
				return fmt.Errorf("failed to update counter %s(%d): %s", m.ID, *m.Delta, err.Error())
			}
		case "gauge":
			_, err = tx.Exec(d.Ctx, sqlGauges, m.ID, m.Value)
			if err != nil {
				return fmt.Errorf("failed to update gauge %s(%f): %s", m.ID, *m.Value, err.Error())
			}
		default:
			// we shuld not be here. Metric type were checked at serializer.DecodeBodyBatch()
			return structs.ErrMetricBadType
		}
	}
	err = tx.Commit(d.Ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %s", err.Error())
	}
	return nil
}

func (d *DBConnector) GetMetrics() ([]structs.Metric, error) {
	var metrics []structs.Metric

	counters, err := d.getAllCounters()
	if err != nil {
		return []structs.Metric{}, fmt.Errorf("cant get counters: %s", err.Error())
	}

	gauges, err := d.getAllGauges()
	if err != nil {
		return []structs.Metric{}, fmt.Errorf("cant get gauges: %s", err.Error())
	}

	for k, v := range counters {
		metrics = append(metrics, structs.Metric{ID: k, Delta: &v, MType: "counter"})
	}

	for k, v := range gauges {
		metrics = append(metrics, structs.Metric{ID: k, Value: &v, MType: "gauge"})
	}

	return metrics, nil
}

func (d *DBConnector) GetMetric(m structs.Metric) (structs.Metric, error) {
	switch m.MType {
	case "counter":
		c, err := d.getCounter(m.ID)
		if err != nil {
			return structs.Metric{}, err
		}
		m.Delta = &c
		return m, nil
	case "gauge":
		g, err := d.getGauge(m.ID)
		if err != nil {
			return structs.Metric{}, err
		}
		m.Value = &g
		return m, nil
	default:
		e := fmt.Errorf("cant get %s. Metric has unknown type: %s", m.ID, m.MType)
		log.Printf("ERROR: %s", e.Error())
		return structs.Metric{}, e
	}
}

func (d *DBConnector) UpdateMetric(m structs.Metric) error {
	switch m.MType {
	case "counter":
		if m.Delta == nil {
			return structs.ErrMetricNullAttr
		}
		err := d.increaseCounter(m.ID, *m.Delta)
		if err != nil {
			return err
		}
	case "gauge":
		if m.Value == nil {
			return structs.ErrMetricNullAttr
		}
		err := d.setGauge(m.ID, *m.Value)
		if err != nil {
			return err
		}
	default:
		log.Printf("WARN: cant get %s. Metric has unknown type: %s", m.ID, m.MType)
		return structs.ErrMetricBadType
	}
	return nil

}
