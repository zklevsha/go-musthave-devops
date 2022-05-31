package structs

type ServerResponse interface {
	AsText() string
	SetHash(key string)
}

type Storage interface {
	GetMetric(metric Metric) (Metric, error)
	GetMetrics() ([]Metric, error)
	UpdateMetric(metric Metric) error
	UpdateMetrics(metrics []Metric) error
	ResetCounter(ID string) error
	Avaliable() error
	Close()
	Init() error
}
