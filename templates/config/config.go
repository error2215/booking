package config

import "time"

const (
	ELASTIC_SERVICE_IP = "http://localhost:9200"
	MAIN_INDEX         = "booking_index"
)

type Order struct {
	Id     int           `json:"-"`
	Author string        `json:"author"`
	Time   time.Duration `json:"time"`
}
