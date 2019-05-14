package main

import (
	"strconv"

	"github.com/spf13/viper"
	"gockets/routes"
	"gockets/setup"
	"gockets/src/services/logger"
)

func main() {
	setup.Init()

	router := routes.InitRoutes()
	port := ":" + strconv.Itoa(viper.GetInt("port"))
	ll.Log.Infof("Server started on %s", port)

	err := router.Run(port)
	if err != nil {
		ll.Log.Fatal(err)
	}
}
