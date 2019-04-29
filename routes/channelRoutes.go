package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	"gockets/helpers"
	"gockets/src/controllers"
	"gockets/src/services/logger"
	"strings"
)

func InitRoutes(hostName string, port int) *mux.Router {
	router := mux.NewRouter()
	initChannelRoutes(router, hostName, port)
	ll.Log.Debug("Routes initialized")
	return router
}

func initChannelRoutes(r *mux.Router, hostName string, port int) {
	fullHostname := fmt.Sprintf("{hostname:%s}:%d", createHostnameRegex(hostName), port)

	cc := r.Host(fullHostname).Subrouter()
	cc.HandleFunc("/channel/prepare", controllers.PrepareChannel).Methods("POST")
	cc.HandleFunc("/channel", controllers.GetAllChannels).Methods("GET")
	cc.HandleFunc("/channel/{publisherToken}", controllers.GetChannel).Methods("GET")
	cc.HandleFunc("/channel/publish/{publisherToken}", controllers.PushToConnection).Methods("POST")
	cc.HandleFunc("/channel/publish/{publisherToken}", controllers.CloseConnection).Methods("DELETE")
	ll.Log.Debugf("Locked administrative routes to access from %s")

	r.HandleFunc("/channel/subscribe/{subscriberToken}", controllers.CreateConnection)
}

func createHostnameRegex(hostname string) string {
	var hostnames = []string{
		"localhost",
	}
	if  hostname != "localhost" {
		hostnames = append(hostnames, hostname)
	}

	var regexHostnames []string

	for _, h := range hostnames {
		ip, err := helpers.LookupName(h)
		if err != nil {
			continue
		}
		regexHostnames = append(regexHostnames, fmt.Sprintf("[%s]+", ip))
		regexHostnames = append(regexHostnames, fmt.Sprintf("[%s]+", h))
	}

	return strings.Join(regexHostnames, "|")
}
