package main

import (
	"gockets/routes"
	"gockets/setup"
	"log"
	"net/http"
	"strconv"
)

func main() {
	setupObject := setup.Setup()
	router := routes.InitRoutes()
	port := ":" + strconv.Itoa(setupObject.Port)
	log.Println("Server started on " + port)
	log.Fatal(http.ListenAndServe(port, router))
}
