package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"websocket/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var PublisherChannels = make(map[string]*models.Channel)

var SubscriberChannels = make(map[string]*models.Channel)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/channel/prepare", PrepareChannel).Methods("GET")
	router.HandleFunc("/channel", GetAllChannels).Methods("GET")
	router.HandleFunc("/channel/subscribe/{subscriberToken}", CreateConnection)
	router.HandleFunc("/channel/publish/{publisherToken}", PushToConnection).Methods("POST")

	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(":8844", router))
}

func GenerateChannel() models.Channel {
	hasher := md5.New()
	timeString := strconv.FormatInt(time.Now().Unix(), 10)
	hasher.Write([]byte(timeString))
	publisherKey := hex.EncodeToString(hasher.Sum(nil))

	hasher.Write([]byte(publisherKey))
	subscriberKey := hex.EncodeToString(hasher.Sum(nil))

	return models.Channel{
		PublisherToken:  publisherKey,
		SubscriberToken: subscriberKey,
	}
}

func PrepareChannel(w http.ResponseWriter, r *http.Request) {
	var channel models.Channel
	for {
		channel = GenerateChannel()
		if _, ok := PublisherChannels[channel.PublisherToken]; ok {
			continue
		} else {
			PublisherChannels[channel.PublisherToken] = &channel
			SubscriberChannels[channel.SubscriberToken] = &channel
			break
		}
	}

	preparedJson, _ := json.Marshal(channel)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}

func GetAllChannels(w http.ResponseWriter, r *http.Request) {
	var allChannels []models.Channel
	for _, value := range PublisherChannels {
		allChannels = append(allChannels, *value)
	}

	preparedJson, _ := json.Marshal(models.Channels {
		Channels: allChannels,
	})

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if channel, ok := SubscriberChannels[vars["subscriberToken"]]; ok {
		socketConnection, _ := upgrader.Upgrade(w, r, nil)
		dataChannel := make(chan string)
		channel.SubscriberChannel = dataChannel
		go PushDataToConnection(socketConnection, dataChannel)
	} else {
		preparedJson, _ := json.Marshal(models.Response{
			Message: "Subscriber token not found",
			Type:    "ERR",
		})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(preparedJson)
	}
}

func PushDataToConnection(socket *websocket.Conn, dataChan chan string) {
	for {
		select {
		case data := <-dataChan:
			_ = socket.WriteMessage(websocket.TextMessage, []byte(data))
		}
	}
}

func PushToConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var response models.Response
	if channel, ok := PublisherChannels[vars["publisherToken"]]; ok {
		if channel.SubscriberChannel == nil {
			response = models.Response{
				Message: "Subscriber has not subscribed yet",
				Type:    "INF",
			}
		} else {
			channel.SubscriberChannel <- r.FormValue("data")
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
	}

	preparedJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}
