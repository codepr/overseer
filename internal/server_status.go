package internal

import (
	"time"
)

type URL = string

type ServerStatus struct {
	Url             URL           `json:"url"`
	Alive           bool          `json:"alive"`
	ResponseTime    time.Duration `json:"response_time"`
	ResponseStatus  int           `json:"response_status"`
	ResponseContent string        `json:"response_content"`
}
