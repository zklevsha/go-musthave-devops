package structs

type ServerResponse interface {
	AsText() string
	SetHash(key string)
}

type Storage interface {
	GetMetric(metric Metric) (Metric, int, error)
	GetMetrics() ([]Metric, int, error)
	UpdateMetric(metric Metric) (int, error)
	UpdateMetrics(metrics []Metric) (int, error)
	ResetCounter(ID string) error
	Avaliable() error
	Close()
	Init() error
}
