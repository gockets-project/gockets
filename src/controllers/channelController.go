package controllers

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gockets/helpers"
	"gockets/models"
	"gockets/src/services/callback"
	"gockets/src/services/channel"
	"gockets/src/services/connection"
	"gockets/src/services/logger"
	"gockets/src/services/tickerHelper"
	"io/ioutil"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CloseConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var response models.Response

	if publisherChannel, ok := channel.PublisherChannels[vars["publisherToken"]]; ok {

		for i := 0; i < publisherChannel.Listeners; i++ {
			publisherChannel.PublisherChannel <- models.ChannelCloseSignal
		}

		delete(channel.PublisherChannels, publisherChannel.PublisherToken)
		delete(channel.SubscriberChannels, publisherChannel.SubscriberToken)

		response = models.Response{
			Message: "Successfully closed connection",
			Type:    "OK",
		}
	} else {
		response = models.Response{
			Message: "Publisher token not found",
			Type:    "ERR",
		}
		w.WriteHeader(http.StatusNotFound)
	}

	helpers.WriteJsonResponse(w, response)
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if subscriberChannel, ok := channel.SubscriberChannels[vars["subscriberToken"]]; ok {
		ll.Log.Debug("Establishing WS connection")
		conn, _ := upgrader.Upgrade(w, r, nil)
		ll.Log.Debug("Connection upgraded")
		ccc := make(chan int)
		conn.SetCloseHandler(func(code int, text string) error {
			switch code {
			case websocket.CloseGoingAway:
				ll.Log.Debug("Client going away (1001)")
				break
			case websocket.CloseNormalClosure:
				ll.Log.Debug("Regular shutdown (1000)")
				break
			case websocket.CloseNoStatusReceived:
				ll.Log.Debug("No status shutdown (1005)")
				break
			default:
				ll.Log.Debugf("Shutdown of connection with code: %d", code)
				break
				// pass to socket connection shutdown signal
			}
			ll.Log.Infof("Sending shutdown signal. Shutdown code: %d", code)

			// signal 1 is for client-side shutdown
			ccc <- 1

			return nil
		})
		ll.Log.Debug("Created CLOSE handler")

		conn.SetPongHandler(func(appData string) error {
			ll.Log.Debug("Pong handler triggered")
			deadline := tickerHelper.GetPingDeadline().Add(time.Duration(2 * time.Second))
			_ = conn.SetReadDeadline(deadline)
			ll.Log.Debug("Set read deadline to %s", deadline.String())
			return nil
		})
		ll.Log.Debug("Created PONG handler")

		subscriberChannel.Listeners++
		ll.Log.Info("Creating WS handle routines")
		ll.Log.Debug("Started PUSH routine")
		go connection.PushDataToConnection(conn, subscriberChannel, ccc)
		ll.Log.Debug("Started CALLBACK routine")
		go callback.HandleSentData(subscriberChannel)
	} else {
		helpers.WriteJsonResponse(w, models.Response{
			Message: "Subscriber token not found",
			Type:    "ERR",
		})
	}
}

func PushToConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var response models.Response
	if publisherChannel, ok := channel.PublisherChannels[vars["publisherToken"]]; ok {
		if publisherChannel.SubscriberChannel == nil {
			response = models.Response{
				Message: "Subscriber has not subscribed yet",
				Type:    "INF",
			}
		} else {
			body, _ := ioutil.ReadAll(r.Body)
			for i := 0; i < publisherChannel.Listeners; i++ {
				publisherChannel.SubscriberChannel <- string(body)
			}
			response = models.Response{
				Message: "Successfully pushed data to subscriber",
				Type:    "INF",
			}
		}
	} else {
		response = models.Response{
			Message: "Publisher token not found",
			Type:    "ERR",
		}
		w.WriteHeader(http.StatusNotFound)
	}
	helpers.WriteJsonResponse(w, response)
}
