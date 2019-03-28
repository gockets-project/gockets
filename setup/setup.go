package setup

import (
	"flag"
	"gockets/models"
	"gockets/src/services/logger"
	"gockets/src/services/tickerHelper"
)

func Setup() models.Setup {
	setupObject := models.Setup{}
	flagPass(&setupObject)
	initPackages(setupObject)
	ll.InitLogger(setupObject.LogLevel)
	ll.Log.Debug("Setup complete")
	return setupObject
}

func flagPass(setupObject *models.Setup) {
	portIntPtr := flag.Int("port", 8844, "Port of a server")
	pingIntPtr := flag.Int("ping-interval", 60, "Interval of ping request and time for pong response for clients in seconds")
	logLvlPtr := flag.Int("log-level", 1, "Level of logging. 1 - Info and error. 2 - Error only. 3 - All info with debug")
	llLockPtr := flag.String("host-name", "localhost", "Lock access to all administrative routes only to access from specific hostname")
	if *logLvlPtr < 1 || *logLvlPtr > 3 {
		panic("Invalid argument passed to log-level")
	}
	flag.Parse()
	setupObject.Port = *portIntPtr
	setupObject.PingDelay = *pingIntPtr
	setupObject.LogLevel = *logLvlPtr
	setupObject.AdminHostname = *llLockPtr
}

func initPackages(setupObject models.Setup) {
	tickerHelper.SetPingInterval(setupObject.PingDelay)
}
