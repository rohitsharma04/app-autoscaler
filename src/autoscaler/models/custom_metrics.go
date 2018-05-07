package models

type CustomMetric struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Timestamp  int64   `json:"timestamp"`
	InstanceID uint32  `json:"instance_index"`
	AppGUID    string  `json:"app_guid"`
}
