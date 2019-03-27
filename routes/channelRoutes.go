package routes

import (
	"github.com/gorilla/mux"
	"gockets/src/controllers"
	"gockets/src/services/channel"
	"gockets/src/services/logger"
)

func InitRoutes(lhRestricted bool) *mux.Router {
	router := mux.NewRouter()
	initChannelRoutes(router, lhRestricted)
	ll.Log.Debug("Routes initialized")
	return router
}

func initChannelRoutes(r *mux.Router, lhRestricted bool) {
	r.PathPrefix("channel")
	if lhRestricted {
		cc := r.Host("localhost").Subrouter()
		cc.HandleFunc("/channel/prepare", channel.PrepareChannel).Methods("POST")
		cc.HandleFunc("/channel", channel.GetAllChannels).Methods("GET")
		cc.HandleFunc("/channel/{publisherToken}", channel.GetChannel).Methods("GET")
		cc.HandleFunc("/channel/publish/{publisherToken}", controllers.PushToConnection).Methods("POST")
		cc.HandleFunc("/channel/publish/{publisherToken}", controllers.CloseConnection).Methods("DELETE")
		ll.Log.Debug("Locked administrative routes to LOCALHOST-ONLY access")
	} else {
		r.HandleFunc("/channel/prepare", channel.PrepareChannel).Methods("POST")
		r.HandleFunc("/channel", channel.GetAllChannels).Methods("GET")
		r.HandleFunc("/channel/{publisherToken}", channel.GetChannel).Methods("GET")
		r.HandleFunc("/channel/subscribe/{subscriberToken}", controllers.CreateConnection)
		r.HandleFunc("/channel/publish/{publisherToken}", controllers.PushToConnection).Methods("POST")
		r.HandleFunc("/channel/publish/{publisherToken}", controllers.CloseConnection).Methods("DELETE")
	}

	r.HandleFunc("/channel/subscribe/{subscriberToken}", controllers.CreateConnection)
}
