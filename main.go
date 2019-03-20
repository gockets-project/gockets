package main

import (
	"gockets/routes"
	"log"
	"net/http"
)

func main() {
	router := routes.InitRoutes()
	log.Println("Server started")
	log.Fatal(http.ListenAndServe(":8844", router))
}
