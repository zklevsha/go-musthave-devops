package db

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/zklevsha/go-musthave-devops/internal/structs"
	"golang.org/x/exp/slices"
)

const dsn = "postgres://go-server:pgdbpwd@localhost:5532/go-musthave-devops_test"

var ctx = context.Background()

func tableExists(tname string, d DBConnector) (bool, error) {
	conn, err := d.Pool.Acquire(d.Ctx)
	if err != nil {
		return false, fmt.Errorf("failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	var count int
	sql := `SELECT COUNT(table_name) FROM information_schema.tables WHERE table_name=$1;`
	row := conn.QueryRow(d.Ctx, sql, tname)
	err = row.Scan(&count)
	if err != nil {
		return false, fmt.Errorf("sql %s have returned an error: %s", sql, err.Error())
	}

	switch count {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("sql %s have returned more than 1 COUNT: %d", sql, count)
	}
}
func initDB() DBConnector {
	d := DBConnector{Ctx: ctx, DSN: dsn}
	err := d.Init()
	if err != nil {
		log.Fatalf("Failed to init DBConnector: %s", err.Error())
	}
	err = d.DropTables()
	if err != nil {
		log.Fatalf("DropTables have returned an error to init DBConnector: %s", err.Error())
	}
	err = d.CreateTables()
	if err != nil {
		log.Fatalf("CreateTables have returned ad error: %s", err.Error())
	}
	return d
}

func TestCheckInit(t *testing.T) {
	notInitilized := DBConnector{Ctx: ctx, DSN: dsn}
	initilized := DBConnector{Ctx: ctx, DSN: dsn}
	err := initilized.Init()
	if err != nil {
		log.Fatalf("Init() have returned an error: %s", err.Error())
	}
	defer initilized.Close()

	tt := []struct {
		name string
		want error
		db   DBConnector
	}{
		{
			name: "not initialized",
			want: fmt.Errorf("DbConnector is not initiliazed (run DBConnector.Init() to initilize"),
			db:   notInitilized,
		},
		{
			name: "initialized",
			want: nil,
			db:   initilized,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.db.checkInit()

			if reflect.TypeOf(res) != reflect.TypeOf(tc.want) {
				t.Errorf("CheckInit() got bad return: have: %v,  want: %v", res, tc.want)
			}
		})
	}

}

func TestInit(t *testing.T) {
	//setup
	goodDSN := DBConnector{Ctx: ctx, DSN: dsn}
	badDSN := DBConnector{Ctx: ctx, DSN: "bad"}

	tt := []struct {
		name string
		want error
		db   DBConnector
	}{
		{
			name: "good dsn",
			want: nil,
			db:   goodDSN,
		},
		{
			name: "bad dsn",
			want: fmt.Errorf("unable to connect to database:"),
			db:   badDSN,
		},
	}

	// run tests
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.db.Init()
			if reflect.TypeOf(err) != reflect.TypeOf(tc.want) {
				t.Errorf("Init() got bad return: have: %v,  want: %v", err, tc.want)
			}
			if err == nil {
				defer tc.db.Close()
			}
		})
	}

	// teardown
	goodDSN.Init()
	err := goodDSN.DropTables()
	if err != nil {
		t.Errorf("DropTable() have returned an error: %s", err.Error())
	}

}

func TestClose(t *testing.T) {
	d := initDB()
	defer d.Close()

	t.Run("test Close()", func(t *testing.T) {
		d.Close()

		if d.initialized != false {
			t.Errorf("initialized still true")
		}

		_, err := d.Pool.Acquire(d.Ctx)
		if err == nil {
			t.Errorf("Pool is not closed")
		}
	})

}

func TestAvailable(t *testing.T) {
	d := initDB()
	defer d.Close()

	t.Run("test Available()", func(t *testing.T) {
		err := d.Avaliable()
		if err != nil {
			t.Errorf("Available() returned an error: %s", err.Error())
		}
	})

}

func TestGetGauge(t *testing.T) {
	// setup
	type gauge struct {
		id    string
		value float64
	}
	goodMetric := gauge{id: "test_metric", value: 154.33}
	nonExitentMetric := gauge{id: "nx_metric", value: -1}

	d := initDB()
	defer d.Close()

	err := d.setGauge(goodMetric.id, goodMetric.value)
	if err != nil {
		log.Fatalf("setGauge returned an error: %s", err.Error())
	}

	// run test cases
	type want struct {
		gauge gauge
		err   error
	}

	tt := []struct {
		name string
		want want
	}{
		{
			name: "good metric",
			want: want{gauge: goodMetric, err: nil},
		},
		{
			name: "non exitent metirc",
			want: want{gauge: nonExitentMetric, err: structs.ErrMetricNotFound},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := d.getGauge(tc.want.gauge.id)
			if res != tc.want.gauge.value {
				t.Errorf("bad value: want: %f, have: %f", tc.want.gauge.value, res)
			}
			if err != tc.want.err {
				t.Errorf("bad error: want: %v, have: %v", tc.want.err, err)
			}
		})
	}

	//teardown
	err = d.DropTables()
	if err != nil {
		t.Errorf("DropTables returned ad error: %s", err.Error())
	}

}

func TestSetGauge(t *testing.T) {
	// setup
	type gauge struct {
		id    string
		value float64
	}
	testGauge := gauge{id: "test_metric", value: 642.3278}

	d := initDB()
	defer d.Close()

	// run test cases
	tt := []struct {
		name  string
		gauge gauge
	}{
		{
			name:  "setGauge",
			gauge: testGauge,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := d.setGauge(tc.gauge.id, tc.gauge.value)
			if err != nil {
				t.Errorf("setGauge returned an error: %s", err.Error())
			}

			value, err := d.getGauge(tc.gauge.id)
			if err != nil {
				t.Errorf("cant check value: getGauge returned an error: %s", err.Error())
			}

			if value != tc.gauge.value {
				t.Errorf("result mimatch: have: %f, want: %f", value, tc.gauge.value)
			}
		})
	}

	//teardown
	err := d.DropTables()
	if err != nil {
		t.Errorf("DropTables returned ad error: %s", err.Error())
	}

}

func TestGetAllGauges(t *testing.T) {
	// setup
	type gauge struct {
		id    string
		value float64
	}
	testGauges := []gauge{
		{id: "gauge1", value: 0.5},
		{id: "gauge2", value: 30},
	}

	d := initDB()
	defer d.Close()

	for _, g := range testGauges {
		err := d.setGauge(g.id, g.value)
		if err != nil {
			log.Fatalf("Failed to set gauge (id:%s, value: %f): %s", g.id, g.value, err.Error())
		}

	}

	// run test
	t.Run("getAllGauges", func(t *testing.T) {
		res, err := d.getAllGauges()
		if err != nil {
			t.Errorf("getAllGauges returned an error: %s", err.Error())
		}
		for _, g := range testGauges {
			if val, ok := res[g.id]; ok {
				if val != g.value {
					t.Errorf("gauge %s mismatch: have %f want %f", g.id, val, g.value)
				}
			} else {
				t.Errorf("gauge %s did not found", g.id)
			}
		}
	})

	//teardown
	err := d.DropTables()
	if err != nil {
		t.Errorf("DropTables returned ad error: %s", err.Error())
	}

}

func TestGetCounter(t *testing.T) {
	// setup
	type counter struct {
		id    string
		value int64
	}
	goodMetric := counter{id: "test_metric", value: 190}
	nonExitentMetric := counter{id: "nx_metric", value: -1}

	d := initDB()
	defer d.Close()

	err := d.increaseCounter(goodMetric.id, goodMetric.value)
	if err != nil {
		log.Fatalf("increaseCounter returned an error: %s", err.Error())
	}

	// run test cases
	type want struct {
		counter counter
		err     error
	}

	tt := []struct {
		name string
		want want
	}{
		{
			name: "good metric",
			want: want{counter: goodMetric, err: nil},
		},
		{
			name: "non exitent metirc",
			want: want{counter: nonExitentMetric, err: structs.ErrMetricNotFound},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := d.getCounter(tc.want.counter.id)
			if res != tc.want.counter.value {
				t.Errorf("bad value: want: %d, have: %d", tc.want.counter.value, res)
			}
			if err != tc.want.err {
				t.Errorf("bad error: want: %v, have: %v", tc.want.err, err)
			}
		})
	}

	//teardown
	err = d.DropTables()
	if err != nil {
		t.Errorf("DropTables returned ad error: %s", err.Error())
	}

}

func TestIncreaseCounter(t *testing.T) {
	// setup
	type counter struct {
		id    string
		value int64
	}
	testCounter := counter{id: "test_metric", value: 100}

	d := initDB()
	defer d.Close()

	// run test cases
	tt := []struct {
		name    string
		counter counter
	}{
		{
			name:    "increaseCounter",
			counter: testCounter,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := d.increaseCounter(tc.counter.id, tc.counter.value)
			if err != nil {
				t.Errorf("setGauge returned an error: %s", err.Error())
			}

			value, err := d.getCounter(tc.counter.id)
			if err != nil {
				t.Errorf("cant check value: getCounter returned an error: %s", err.Error())
			}

			if value != tc.counter.value {
				t.Errorf("result mimatch: have: %d, want: %d", value, tc.counter.value)
			}
		})
	}

	//teardown
	err := d.DropTables()
	if err != nil {
		t.Errorf("DropTables returned ad error: %s", err.Error())
	}

}

func TestResetCounter(t *testing.T) {
	// setup
	type counter struct {
		id    string
		value int64
	}
	testCounter := counter{id: "test_metric", value: 100}

	d := initDB()
	defer d.Close()

	err := d.increaseCounter(testCounter.id, testCounter.value)
	if err != nil {
		log.Fatalf("increaseCounter returned an error: %s", err.Error())
	}

	// run
	t.Run("resetCounter", func(t *testing.T) {
		err := d.ResetCounter(testCounter.id)
		if err != nil {
			t.Errorf("resetCounter have returned an error: %s", err.Error())
		}

		value, err := d.getCounter(testCounter.id)
		if err != nil {
			t.Errorf("cant check value: testCounter have returned an error: %s", err.Error())
		}

		if value != 0 {
			t.Errorf("result mimatch: have: %d, want: %d", value, 0)
		}
	})

	//teardown
	err = d.DropTables()
	if err != nil {
		t.Errorf("DropTables have returned ad error: %s", err.Error())
	}

}

func TestCreateTable(t *testing.T) {
	// setup
	d := initDB()
	defer d.Close()
	d.DropTables()

	// run tests
	t.Run("CreateTables", func(t *testing.T) {
		err := d.CreateTables()
		if err != nil {
			t.Errorf("CreateTables have returned an error: %s", err.Error())
		}

		gaugesExist, err := tableExists("gauges", d)
		if err != nil {
			t.Errorf("tableExists() have returned an error: %s", err.Error())
		}
		if !gaugesExist {
			t.Errorf("table 'gauges' was not created")
		}

		countersExist, err := tableExists("counters", d)
		if err != nil {
			t.Errorf("tableExists() have returned an error: %s", err.Error())
		}
		if !countersExist {
			t.Errorf("table 'counters' was not created")
		}
	})

	// teardown
	err := d.DropTables()
	if err != nil {
		t.Errorf("DropTables have returned ad error: %s", err.Error())
	}
}

func TestDropTable(t *testing.T) {
	// setup
	d := initDB()
	defer d.Close()

	// run tests
	t.Run("DropTables", func(t *testing.T) {
		err := d.DropTables()
		if err != nil {
			t.Errorf("DropTables have returned an error: %s", err.Error())
		}

		gauges_exists, err := tableExists("gauges", d)
		if err != nil {
			t.Errorf("tableExists() have returned an error: %s", err.Error())
		}
		if gauges_exists {
			t.Errorf("table 'gauges' still exists")
		}

		counters_exist, err := tableExists("counters", d)
		if err != nil {
			t.Errorf("tableExists() have returned an error: %s", err.Error())
		}
		if counters_exist {
			t.Errorf("table 'counters' still exists")
		}
	})

	// teardown
	err := d.DropTables()
	if err != nil {
		t.Errorf("DropTables have returned ad error: %s", err.Error())
	}
}

func TestUpdateMetrics(t *testing.T) {
	// setup
	d := initDB()
	defer d.Close()

	var gaugeValue = 13.120044
	var counterValue int64 = 100

	testMetrics := []structs.Metric{
		{ID: "test_gauge1", Value: &gaugeValue, MType: "gauge"},
		{ID: "test_counter1", Delta: &counterValue, MType: "counter"},
	}

	// run tests
	t.Run("UpdateMetrics", func(t *testing.T) {
		err := d.UpdateMetrics(testMetrics)
		if err != nil {
			t.Logf("UpdateMetrics() have returned an error: %s", err.Error())
			t.FailNow()
		}
		for _, m := range testMetrics {
			res, err := d.GetMetric(m)
			if err != nil {
				t.Errorf("cant get metric %s, GetMetric() have returned an error: %s", m.ID, err.Error())
			}
			if res.MType != m.MType {
				t.Errorf("Metrics type mismatch: have %s, want: %s", res.MType, m.MType)
			}
			if m.MType == "gauge" && *m.Value != *res.Value {
				t.Errorf("Metrics Value mismatch: have: %f, want %f", *res.Value, *m.Value)
			}
			if m.MType == "counter" && *m.Delta != *res.Delta {
				t.Errorf("Metrics Delta mismatch: have: %d, want: %d", *res.Delta, *m.Delta)
			}
		}
	})

	// teardown
	err := d.DropTables()
	if err != nil {
		t.Errorf("DropTables have returned ad error: %s", err.Error())
	}
}

func TestGetMetrics(t *testing.T) {
	// setup
	d := initDB()
	defer d.Close()

	var gaugeValue = 1134223.006574
	var counterValue int64 = 100

	testMetrics := []structs.Metric{
		{ID: "test_gauge1", Value: &gaugeValue, MType: "gauge"},
		{ID: "test_counter1", Delta: &counterValue, MType: "counter"},
	}
	err := d.UpdateMetrics(testMetrics)
	if err != nil {
		log.Fatalf("UpdateMetrics have returned an error: %s", err.Error())
	}

	// run tests
	t.Run("GetMetrics", func(t *testing.T) {
		metrics, err := d.GetMetrics()
		if err != nil {
			t.Logf("GetMetrics() have returned an error: %s", err.Error())
			t.FailNow()
		}
		for _, tm := range testMetrics {
			idx := slices.IndexFunc(metrics, func(m structs.Metric) bool { return m.ID == tm.ID })
			if idx == -1 {
				t.Errorf("Metic %s does not exists in received Metrics list %v", tm.ID, metrics)
				continue
			}
			if metrics[idx].MType != tm.MType {
				t.Errorf("Metrics type mismatch: have: %s, want: %s", metrics[idx].MType, tm.MType)
			}
			if tm.MType == "gauge" && *tm.Value != *metrics[idx].Value {
				t.Errorf("Metric Value mismatch: have: %f, want: %f", *metrics[idx].Value, *tm.Value)
			}
			if tm.MType == "counter" && *tm.Delta != *metrics[idx].Delta {
				t.Errorf("Metric Value mismatch: have: %d, want: %d", *metrics[idx].Delta, *tm.Delta)
			}
		}
	})

	// teardown
	err = d.DropTables()
	if err != nil {
		t.Errorf("DropTables have returned ad error: %s", err.Error())
	}
}

func TestGetMetric(t *testing.T) {
	// setup
	d := initDB()
	defer d.Close()

	var gaugeValue = 1134223.006574
	var counterValue int64 = 100

	testMetrics := []structs.Metric{
		{ID: "test_gauge1", Value: &gaugeValue, MType: "gauge"},
		{ID: "test_counter1", Delta: &counterValue, MType: "counter"},
	}
	err := d.UpdateMetrics(testMetrics)
	if err != nil {
		log.Fatalf("UpdateMetrics have returned an error: %s", err.Error())
	}

	// run tests
	t.Run("GetMetric", func(t *testing.T) {
		for _, m := range testMetrics {
			res, err := d.GetMetric(m)
			if err != nil {
				t.Errorf("GetMetric() have returned an error: %s", err.Error())
			}
			if res.MType != m.MType {
				t.Errorf("Metric Type mismatch: have: %s, want: %s", res.MType, m.MType)
				continue
			}
			if m.MType == "gauge" && *m.Value != *res.Value {
				t.Errorf("Metric Value mismatch: have: %f, want:%f", *res.Value, *m.Value)
			}
			if m.MType == "counter" && *m.Delta != *res.Delta {
				t.Errorf("Metric Value mismatch: have: %d, want:%d", *res.Delta, *m.Delta)
			}
		}
	})

	// teardown
	err = d.DropTables()
	if err != nil {
		t.Errorf("DropTables have returned ad error: %s", err.Error())
	}
}

func TestUpdateMetric(t *testing.T) {
	// setup
	d := initDB()
	defer d.Close()

	var gaugeValue = 1134223.006574
	var counterDelta int64 = 100

	var newGaugeValue = 0.1
	var newCounterDelta int64 = 1

	var deltaWant = counterDelta + newCounterDelta

	testMetrics := []structs.Metric{
		{ID: "test_gauge1", Value: &gaugeValue, MType: "gauge"},
		{ID: "test_counter1", Delta: &counterDelta, MType: "counter"},
	}

	newMetrics := []structs.Metric{
		{ID: "test_gauge1", Value: &newGaugeValue, MType: "gauge"},
		{ID: "test_counter1", Delta: &newCounterDelta, MType: "counter"},
	}

	err := d.UpdateMetrics(testMetrics)
	if err != nil {
		log.Fatalf("UpdateMetrics have returned an error: %s", err.Error())
	}

	// run tests
	t.Run("UpdateMetric", func(t *testing.T) {
		for _, m := range newMetrics {
			err := d.UpdateMetric(m)
			if err != nil {
				t.Errorf("UpdateMetric() have returned an error: %s", err.Error())
				continue
			}
			res, err := d.GetMetric(m)
			if err != nil {
				t.Errorf("GetMetric() have returned an error: %s", err.Error())
				continue
			}
			if res.MType != m.MType {
				t.Errorf("Metric Type mismatch: have: %s, want: %s", res.MType, m.MType)
				continue
			}
			if m.MType == "gauge" && *m.Value != *res.Value {
				t.Errorf("Metric Value mismatch: have: %f, want:%f", *res.Value, *m.Value)
			}
			if m.MType == "counter" && *res.Delta != deltaWant {
				t.Errorf("Metric Value mismatch: have: %d, want:%d", *res.Delta, deltaWant)
			}
		}
	})

	// teardown
	err = d.DropTables()
	if err != nil {
		t.Errorf("DropTables have returned ad error: %s", err.Error())
	}
}
