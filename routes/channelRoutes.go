package routes

import (
	"github.com/gorilla/mux"
	"gockets/src/channel"
	"gockets/src/connection"
)

func InitRoutes() *mux.Router {
	router := mux.NewRouter()
	initChannelRoutes(router)

	return router
}

func initChannelRoutes(r *mux.Router) {
	r.HandleFunc("/channel/prepare", channel.PrepareChannel).Methods("POST")
	r.HandleFunc("/channel", channel.GetAllChannels).Methods("GET")
	r.HandleFunc("/channel/subscribe/{subscriberToken}", connection.CreateConnection)
	r.HandleFunc("/channel/publish/{publisherToken}", connection.PushToConnection).Methods("POST")
	r.HandleFunc("/channel/publish/{publisherToken}", connection.CloseConnection).Methods("DELETE")
}