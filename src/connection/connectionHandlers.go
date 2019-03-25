package connection

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gockets/models"
	"gockets/src/callback"
	"gockets/src/channel"
	"gockets/src/tickerHelper"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if subscriberChannel, ok := channel.SubscriberChannels[vars["subscriberToken"]]; ok {
		socketConnection, _ := upgrader.Upgrade(w, r, nil)
		socketConnection.SetCloseHandler(func(code int, text string, ) error {
			switch code {
			case websocket.CloseGoingAway:
				log.Print("Client going away (1001)")
				break
			case websocket.CloseNormalClosure:
				log.Print("Regular shutdown (1000)")
				break
			case websocket.CloseNoStatusReceived:
				log.Print("No status shutdown (1005)")
				break
			default:
				log.Print("Shutdown of connection with code: " + strconv.Itoa(code))
			}

			_ = socketConnection.Close()
			log.Print("Corrupting connection to prevent further reads.")

			return nil
		})
		socketConnection.SetPongHandler(func(appData string) error {
			log.Print("Pong handler triggered")
			dealine := tickerHelper.GetPingDeadline().Add(time.Duration(2 * time.Second))
			_ = socketConnection.SetReadDeadline(dealine)
			log.Print("Set read deadline to" + dealine.String())
			return nil
		})
		subscriberChannel.Listeners++

		go pushDataToConnection(socketConnection, subscriberChannel)
		go readDataFromConnection(socketConnection, subscriberChannel)
		go callback.HandleSentData(subscriberChannel)
	} else {
		preparedJson, _ := json.Marshal(models.Response{
			Message: "Subscriber token not found",
			Type:    "ERR",
		})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(preparedJson)
	}
}

func readDataFromConnection(socket *websocket.Conn, channel *models.Channel) {
	for {
		messageType, p, err := socket.ReadMessage()
		if err != nil {
			log.Println(err)
			log.Println(messageType)
			log.Println("Error caught. Closing socket")
			_ = socket.Close()
			return
		}

		switch messageType {
		case websocket.TextMessage:
			log.Println("Got a text message from listener")
			channel.SubscriberMessagesChannel <- string(p)
			break
		default:
			log.Println("Unsupported message received. Ignoring...")
			break
		}
	}
}

func pushDataToConnection(socket *websocket.Conn, channel *models.Channel) {
	log.Print("Routine started")
	tickerChan := tickerHelper.RunTicker()
	log.Print("Ticker started")
	for {
		select {
		case data := <-channel.SubscriberChannel:
			log.Print("Received data: " + data)
			_ = socket.WriteMessage(websocket.TextMessage, []byte(data))

			log.Print("Data sent")
			break
		case signal := <-channel.PublisherChannel:
			switch signal {
			case models.ChannelCloseSignal:
				log.Print("Received close signal")
				_ = socket.WriteControl(websocket.CloseMessage, []byte{}, tickerHelper.GetPingDeadline())
				_ = socket.Close()
				channel.ResponseChannel <- models.ChannelSignalOk
				log.Print("Routine closed")
				return
			}
			break
		case tickerTime := <-tickerChan.C:
			log.Print("Got ticker push: " + tickerTime.String())
			log.Print("Writing ping message")
			log.Print(time.Now().String())
			log.Print(tickerHelper.GetPingDeadline().String())
			err := socket.WriteControl(websocket.PingMessage, []byte{}, tickerHelper.GetPingDeadline())
			if err != nil {
				log.Print("Cannot ping client")
				log.Print(err)
				log.Print("Removing listener")
				channel.Listeners--
				log.Print("Closing routine")
				return
			}
			log.Print("Wrote ping message")
		}
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
	}

	preparedJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
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
	}

	preparedJson, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}
