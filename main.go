package main

import (
	"bytes"
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
	router.HandleFunc("/channel/prepare", PrepareChannel).Methods("POST")
	router.HandleFunc("/channel", GetAllChannels).Methods("GET")
	router.HandleFunc("/channel/subscribe/{subscriberToken}", CreateConnection)
	router.HandleFunc("/channel/publish/{publisherToken}", PushToConnection).Methods("POST")
	router.HandleFunc("/channel/publish/{publisherToken}", CloseConnection).Methods("DELETE")

	fmt.Println("Server started")
	log.Fatal(http.ListenAndServe(":8844", router))
}

func GenerateChannel(r *http.Request) models.Channel {

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

func PrepareChannel(w http.ResponseWriter, r *http.Request) {
	var channel models.Channel
	for {
		channel = GenerateChannel(r)
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

	preparedJson, _ := json.Marshal(models.Channels{
		Channels: allChannels,
	})

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if channel, ok := SubscriberChannels[vars["subscriberToken"]]; ok {
		socketConnection, _ := upgrader.Upgrade(w, r, nil)
		channel.Listeners++

		go PushDataToConnection(socketConnection, channel)
		go ReadDataFromConnection(socketConnection, channel)
		go HandleSentData(channel)
	} else {
		preparedJson, _ := json.Marshal(models.Response{
			Message: "Subscriber token not found",
			Type:    "ERR",
		})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(preparedJson)
	}
}

func ReadDataFromConnection(socket *websocket.Conn, channel *models.Channel) {
	for {
		messageType, p, err := socket.ReadMessage()
		if err != nil {
			log.Println(err)
			log.Println(messageType)
			log.Println("Error caught. Removing a listener")
			channel.Listeners--
			_ = socket.Close()
			return
		}

		switch messageType {
		case websocket.TextMessage:
			log.Println("Got a text message from listener")
			channel.SubscriberMessagesChannel <- string(p)
			break
		case websocket.CloseMessage:
			log.Println("Listener closed connection")
			channel.Listeners--
			_ = socket.Close()
		}
	}
}

func HandleSentData(channel *models.Channel) {
	for {
		select {
		case message := <-channel.SubscriberMessagesChannel:
			go SendDataToHook(channel, message)
			break
		}
	}
}

func SendDataToHook(channel *models.Channel, data string) {
	req, _ := http.NewRequest("POST", channel.SubscriberMessageHookUrl, bytes.NewBuffer([]byte(data)))
	client := &http.Client{}
	_, _ = client.Do(req)
	return
}

func PushDataToConnection(socket *websocket.Conn, channel *models.Channel) {
	for {
		select {
		case data := <-channel.SubscriberChannel:
			_ = socket.WriteMessage(websocket.TextMessage, []byte(data))
			break
		case signal := <-channel.PublisherChannel:
			switch signal {
			case models.ChannelCloseSignal:
				_ = socket.Close()
				channel.ResponseChannel <- models.ChannelSignalOk
			}
			break
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
			for i := 0; i < channel.Listeners; i++ {
				channel.SubscriberChannel <- r.FormValue("data")
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
	}

	preparedJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}

func CloseConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var response models.Response

	if channel, ok := PublisherChannels[vars["publisherToken"]]; ok {

		for i := 0; i < channel.Listeners; i++ {
			channel.PublisherChannel <- models.ChannelCloseSignal
		}

		delete(PublisherChannels, channel.PublisherToken)
		delete(SubscriberChannels, channel.SubscriberToken)

		response = models.Response{
			Message: "Successfully closed connection",
			Type:    "OK",
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
