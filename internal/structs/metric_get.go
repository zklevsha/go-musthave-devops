package structs

type MetricGet struct {
	ID    string `json:"id" example:"CPU"`     // имя метрики
	MType string `json:"type" example:"gauge"` // параметр, принимающий значение gauge или counter
}
