package connection

import (
	"github.com/gorilla/websocket"
	"gockets/models"
	"gockets/src/services/logger"
	"gockets/src/services/tickerHelper"
)

func readDataFromConnection(socket *websocket.Conn, channel *models.Channel, pcc chan int) {
	for {
		messageType, p, err := socket.ReadMessage()
		if err != nil {
			ll.Log.Errorf("Reading error caught: %s. Closing READ routine.", err)
			pcc <- 1
			return
		}

		switch messageType {
		case websocket.TextMessage:
			ll.Log.Infof("Got a text message from listener: %s", string(p))
			channel.SubscriberMessagesChannel <- string(p)
			break
		default:
			ll.Log.Debugf("Got unsupported message. Ignoring...")
			break
		}
	}
}

func PushDataToConnection(conn *websocket.Conn, ch *models.Channel, ccc chan int) {
	tickerChan := tickerHelper.RunTicker()
	ll.Log.Debug("PUSH routine ticker started")
	pcc := make(chan int)
	go readDataFromConnection(conn, ch, pcc)
	ll.Log.Debug("Created PUSH routine")
	for {
		select {
		case data := <-ch.SubscriberChannel:
			ll.Log.Infof("PUSH routine received data: %s", data)
			_ = conn.WriteMessage(websocket.TextMessage, []byte(data))
			ll.Log.Debug("PUSH routine successfully posted data")
			break
		case signal := <-ch.PublisherChannel:
			switch signal {
			case models.ChannelCloseSignal:
				ll.Log.Info("PUSH routine received CLOSE signal")
				_ = conn.WriteControl(websocket.CloseMessage, []byte{}, tickerHelper.GetPingDeadline())
				_ = conn.Close()
				ch.ResponseChannel <- models.ChannelSignalOk
				ll.Log.Info("PUSH routine closed")
				return
			}
			break
		case <-tickerChan.C:
			ll.Log.Debug("Got ticker push")
			ll.Log.Debug("Writing ping message")
			err := conn.WriteControl(websocket.PingMessage, []byte{}, tickerHelper.GetPingDeadline())
			if err != nil {
				ll.Log.Info("Cannot ping client. Considering it disconnected.")
				handleConnClose(conn, ch)
				ll.Log.Info("Closing PUSH routine")
				return
			}
			ll.Log.Debug("Wrote ping message successfully")
			break
		case <-ccc:
			ll.Log.Info("PUSH routine got client disconnect message. Closing PUSH routine")
			handleConnClose(conn, ch)
			return
		case <-pcc:
			ll.Log.Info("PUSH routine got shutdown message from READ routine. Closing PUSH routine")
			handleConnClose(conn, ch)
			return
		}
	}
}

func handleConnClose(conn *websocket.Conn, ch *models.Channel) {
	_ = conn.Close()
	ch.Listeners--
}
