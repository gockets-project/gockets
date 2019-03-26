package models

type Setup struct {
	Port      int `json:"port"`
	PingDelay int `json:"ping_delay"`
	LogLevel  int `json:"log_level"`
}
