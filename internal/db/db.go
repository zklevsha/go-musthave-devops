package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DBConnector struct {
	DSN        string
	Ctx        context.Context
	Pool       *pgxpool.Pool
	initalized bool
}

func (d *DBConnector) Init(ctx context.Context) error {
	d.Ctx = ctx
	p, err := pgxpool.Connect(d.Ctx, d.DSN)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	d.Pool = p
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
	if !d.initalized {
		err := fmt.Errorf("DbConnector is not initiliazed (run DBConnector.Init() to initilize)")
		return err
	}
	return d.Pool.Ping(d.Ctx)
}

func (d *DBConnector) GetGauge(metricName string) (float64, error) {
	return 1, nil
}

func (d *DBConnector) SetGauge(metricName string, metricValue float64) {

}

func (d *DBConnector) GetAllGauges() map[string]float64 {
	return make(map[string]float64)
}

func (d *DBConnector) GetCounter(metricName string) (int64, error) {
	return 1, nil
}

func (d *DBConnector) SetCounter(metricName string, metricValue int64) {

}
func (d *DBConnector) IncreaseCounter(metricName string, metricValue int64) {

}
func (d *DBConnector) GetAllCounters() map[string]int64 {
	return make(map[string]int64)
}
func (d *DBConnector) ResetCounter(metricName string) error {
	return nil
}

// func (d *DBConnector) CreateTables() error {
// 	con, err := pgx.Connect(d.Ctx, d.DSN)
// 	if err != nil {
// 		return err
// 	}
// 	defer con.Close(d.Ctx)

// 	// countersSQL := `CREATE TABLE IF NOT EXISTS counters(
// 	// 	metric_id varchar(45) NOT NULL,
// 	// 	metric_value integer NOT NULL,
// 	// 	PRIMARY KEY (metric_id)
// 	//   )`

// 	// cgaugesSQL = CREATE TABLE IF NOT EXISTS gauges(
// 	// 	metric_id varchar(45) NOT NULL,
// 	// 	metric_value double precision NOT NULL,
// 	// 	PRIMARY KEY (metric_id)
// 	// )

// }
