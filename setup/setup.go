package setup

import (
	"flag"
	"gockets/models"
	"gockets/src/tickerHelper"
)

func Setup() models.Setup {
	setupObject := models.Setup{}
	flagPass(&setupObject)
	initPackages(setupObject)
	return setupObject
}

func flagPass(setupObject *models.Setup) {
	portIntPtr := flag.Int("port", 8844, "Port of a server")
	pingIntPtr  := flag.Int("ping-interval", 60, "Interval of ping request and time for pong response for clients in seconds")
	flag.Parse()
	setupObject.Port = *portIntPtr
	setupObject.PingDelay = *pingIntPtr
}

func initPackages(setupObject models.Setup) {
	tickerHelper.SetPingInterval(setupObject.PingDelay)
}