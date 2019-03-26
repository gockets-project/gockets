package main

import (
	"gockets/routes"
	"gockets/setup"
	"gockets/src/logger"
	"net/http"
	"strconv"
)

func main() {
	setupObject := setup.Setup()
	router := routes.InitRoutes()
	port := ":" + strconv.Itoa(setupObject.Port)
	ll.Log.Infof("Server started on %s", port)
	err := http.ListenAndServe(port, router)
	if err != nil {
		ll.Log.Fatal(err)
	}
}
