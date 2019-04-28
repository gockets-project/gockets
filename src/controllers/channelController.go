package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gockets/helpers"
	"gockets/models"
	"gockets/src/services/callback"
	"gockets/src/services/connection"
	"gockets/src/services/logger"
	"gockets/src/services/tickerHelper"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var publisherChannels = make(map[string]*models.Channel)

var subscriberChannels = make(map[string]*models.Channel)

func CloseConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var response models.Response
	var code = http.StatusOK

	if publisherChannel, ok := publisherChannels[vars["publisherToken"]]; ok {

		for i := 0; i < publisherChannel.Listeners; i++ {
			publisherChannel.PublisherChannel <- models.ChannelCloseSignal
		}

		delete(publisherChannels, publisherChannel.PublisherToken)
		delete(subscriberChannels, publisherChannel.SubscriberToken)

		response = models.Response{
			Message: "Successfully closed connection",
			Type:    "OK",
		}
	} else {
		response = models.Response{
			Message: "Publisher token not found",
			Type:    "ERR",
		}
		code = http.StatusNotFound
	}

	helpers.WriteJsonResponse(w, response, code)
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if subscriberChannel, ok := subscriberChannels[vars["subscriberToken"]]; ok {
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
		if *subscriberChannel.SubscriberMessageHookUrl != "" {
			ll.Log.Debug("Started CALLBACK routine")
			go callback.HandleSentData(subscriberChannel)
		} else {
			ll.Log.Debug("Ignored CALLBACK routine. No hook url specified")
		}
	} else {
		helpers.WriteJsonResponse(w, models.Response{
			Message: "Subscriber token not found",
			Type:    "ERR",
		}, http.StatusNotFound)
	}
}

func PushToConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var response models.Response
	var code = http.StatusOK

	if publisherChannel, ok := publisherChannels[vars["publisherToken"]]; ok {
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
		code = http.StatusNotFound
	}
	helpers.WriteJsonResponse(w, response, code)
}

func PrepareChannel(w http.ResponseWriter, r *http.Request) {
	ll.Log.Debugf("Channel prepared by: %s", r.Host)
	var channel models.Channel
	for {
		channel = generateChannel(r)
		if _, ok := publisherChannels[channel.PublisherToken]; ok {
			continue
		} else {
			publisherChannels[channel.PublisherToken] = &channel
			subscriberChannels[channel.SubscriberToken] = &channel
			break
		}
	}

	preparedJson, _ := json.Marshal(channel)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}

func GetAllChannels(w http.ResponseWriter, r *http.Request) {
	var allChannels []models.Channel
	for _, value := range publisherChannels {
		allChannels = append(allChannels, *value)
	}

	preparedJson, _ := json.Marshal(models.Channels{
		Channels: allChannels,
	})

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}

func GetChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var preparedJson []byte
	w.Header().Set("Content-Type", "application/json")
	if publisherChannel, ok := publisherChannels[vars["publisherToken"]]; ok {
		preparedJson, _ = json.Marshal(publisherChannel)
	} else {
		preparedJson, _ = json.Marshal(models.Response{
			Message: "Publisher token not found",
			Type:    "ERR",
		})
		w.WriteHeader(http.StatusNotFound)
	}

	_, _ = w.Write(preparedJson)
}

func generateChannel(r *http.Request) models.Channel {

	decoder := json.NewDecoder(r.Body)
	var c models.Channel
	_ = decoder.Decode(&c)

	hasher := md5.New()
	timeString := strconv.FormatInt(time.Now().Unix(), 10)
	hasher.Write([]byte(timeString))
	publisherKey := hex.EncodeToString(hasher.Sum(nil))

	hasher.Write([]byte(publisherKey))
	subscriberKey := hex.EncodeToString(hasher.Sum(nil))

	return models.Channel{
		PublisherToken:           publisherKey,
		SubscriberToken:          subscriberKey,
		SubscriberMessageHookUrl: c.SubscriberMessageHookUrl,
		Listeners:                0,

		ResponseChannel:           make(chan int),
		PublisherChannel:          make(chan int),
		SubscriberChannel:         make(chan string),
		SubscriberMessagesChannel: make(chan string),
	}
}
