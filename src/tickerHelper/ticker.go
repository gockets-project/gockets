package tickerHelper

import "time"

var pingInterval int

func SetPingInterval(intervalSec int) int {
	pingInterval = intervalSec
	return pingInterval
}

func GetPingInterval() int {
	return pingInterval
}

func GetPingDeadline() time.Time {
	return time.Now().Add(time.Duration(pingInterval) * time.Second)
}

func RunTicker() *time.Ticker {
	return time.NewTicker(time.Duration(pingInterval) * time.Second)
}
