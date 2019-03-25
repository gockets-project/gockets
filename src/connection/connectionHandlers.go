package connection

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gockets/helpers"
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
		conn, _ := upgrader.Upgrade(w, r, nil)
		log.Print("Connection upgraded")
		ccc := make(chan int)
		conn.SetCloseHandler(func(code int, text string, ) error {
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
				break
				// pass to socket connection shutdown signal
			}
			log.Print("Sending shutdown signal to pushing routine")

			// signal 1 is for client-side shutdown
			ccc <- 1

			return nil
		})
		log.Print("Created CLOSE handler")
		conn.SetPongHandler(func(appData string) error {
			log.Print("Pong handler triggered")
			deadline := tickerHelper.GetPingDeadline().Add(time.Duration(2 * time.Second))
			_ = conn.SetReadDeadline(deadline)
			// if connection corrupt send shutdown signal
			log.Print("Set read deadline to" + deadline.String())
			return nil
		})

		log.Print("Created PONG handler")
		subscriberChannel.Listeners++
		log.Print("Started PUSH routine")
		go pushDataToConnection(conn, subscriberChannel, ccc)
		log.Print("Started CALLBACK routine")
		go callback.HandleSentData(subscriberChannel)
	} else {
		helpers.WriteJsonResponse(w, models.Response{
			Message: "Subscriber token not found",
			Type:    "ERR",
		})
	}
}

func readDataFromConnection(socket *websocket.Conn, channel *models.Channel, pcc chan int) {
	for {
		log.Print("Reading from socket")
		time.Sleep(1 * time.Second)
		log.Print()
		messageType, p, err := socket.ReadMessage()
		if err != nil {
			log.Println("Reading error caught. Closing READ routine")
			log.Println(err)
			log.Println(messageType)
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

func pushDataToConnection(conn *websocket.Conn, ch *models.Channel, ccc chan int) {
	log.Print("Routine started")
	tickerChan := tickerHelper.RunTicker()
	log.Print("Ticker started")
	pcc := make(chan int)
	log.Print("Ticker started")
	go readDataFromConnection(conn, ch, pcc)
	log.Print("Created READER routine")
	for {
		select {
		case data := <- ch.SubscriberChannel:
			log.Print("Received data: " + data)
			_ = conn.WriteMessage(websocket.TextMessage, []byte(data))

			log.Print("Data sent")
			break
		case signal := <- ch.PublisherChannel:
			switch signal {
			case models.ChannelCloseSignal:
				log.Print("Received close signal")
				_ = conn.WriteControl(websocket.CloseMessage, []byte{}, tickerHelper.GetPingDeadline())
				_ = conn.Close()
				ch.ResponseChannel <- models.ChannelSignalOk
				log.Print("Routine closed")
				return
			}
			break
		case tickerTime := <-tickerChan.C:
			log.Print("Got ticker push: " + tickerTime.String())
			log.Print("Writing ping message")
			log.Print(time.Now().String())
			log.Print(tickerHelper.GetPingDeadline().String())
			err := conn.WriteControl(websocket.PingMessage, []byte{}, tickerHelper.GetPingDeadline())
			if err != nil {
				log.Print("Cannot ping client")
				log.Print(err)
				log.Print("Removing listener")
				handleConnClose(conn, ch)
				log.Print("Closing routine")
				return
			}
			log.Print("Wrote ping message")
			break
		case <- ccc:
			log.Print("Received client-side shutdown signal")
			log.Print("Closing routine")
			handleConnClose(conn, ch)
			return
		}
	}
}

func handleConnClose(conn *websocket.Conn, ch *models.Channel)  {
	_ = conn.Close()
	ch.Listeners--
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
	helpers.WriteJsonResponse(w, response)
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

	helpers.WriteJsonResponse(w, response)
}
