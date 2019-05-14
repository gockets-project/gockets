package controllers

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gockets/models"
	"gockets/src/services/callback"
	"gockets/src/services/connection"
	"gockets/src/services/logger"
	"gockets/src/services/tickerHelper"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var publisherChannels = make(map[string]*models.Channel)

var subscriberChannels = make(map[string]*models.Channel)

func CloseConnection(c *gin.Context) {
	publisherToken := c.Param("publisherToken")

	if publisherChannel, ok := publisherChannels[publisherToken]; ok {

		for i := 0; i < publisherChannel.Listeners; i++ {
			publisherChannel.PublisherChannel <- models.ChannelCloseSignal
		}

		delete(publisherChannels, publisherChannel.PublisherToken)
		delete(subscriberChannels, publisherChannel.SubscriberToken)

		c.JSON(200, models.Response{
			Message: "Successfully closed connection",
			Type:    "OK",
		})
	} else {
		c.JSON(404, models.Response{
			Message: "Publisher token not found",
			Type:    "ERR",
		})
	}
}

func CreateConnection(c *gin.Context) {
	subscriberToken := c.Param("subscriberToken")

	if subscriberChannel, ok := subscriberChannels[subscriberToken]; ok {
		ll.Log.Debug("Establishing WS connection")
		conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
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
			err := conn.SetReadDeadline(deadline)
			ll.Log.Debug("Set read deadline to %s", deadline.String())
			return err
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
		c.JSON(404, models.Response{
			Message: "Subscriber token not found",
			Type:    models.ResponseErr,
		})
	}
}

func PushToConnection(c *gin.Context) {
	publisherToken := c.Param("publisherToken")

	if publisherChannel, ok := publisherChannels[publisherToken]; ok {
		if publisherChannel.SubscriberChannel == nil {
			c.JSON(200, models.Response{
				Message: "Subscriber has not subscribed yet",
				Type:    models.ResponseInf,
			})
		} else {
			body, _ := ioutil.ReadAll(c.Request.Body)
			for i := 0; i < publisherChannel.Listeners; i++ {
				publisherChannel.SubscriberChannel <- string(body)
			}
			c.JSON(200, models.Response{
				Message: "Successfully pushed data to subscriber",
				Type:    models.ResponseInf,
			})
		}
	} else {
		c.JSON(404, models.Response{
			Message: "Publisher token not found",
			Type:    models.ResponseErr,
		})
	}
}

func PrepareChannel(c *gin.Context) {
	ll.Log.Debugf("Channel prepared by: %s", c.Request.Host)
	var channel models.Channel
	for {
		channel = models.GenerateChannel(c.Request)
		if _, ok := publisherChannels[channel.PublisherToken]; ok {
			continue
		} else {
			publisherChannels[channel.PublisherToken] = &channel
			subscriberChannels[channel.SubscriberToken] = &channel
			break
		}
	}

	c.JSON(200, channel)
}

func GetAllChannels(c *gin.Context) {
	var allChannels []models.Channel

	regexString := c.DefaultQuery("search", ".+")

	ll.Log.Debug(regexString)

	r, err := regexp.Compile(regexString)

	if err != nil {
		c.JSON(500, models.Response{
			Type:    models.ResponseErr,
			Message: "Invalid regex pattern was passed",
		})
		return
	}

	for _, value := range publisherChannels {
		if r.MatchString(*value.Tag) {
			allChannels = append(allChannels, *value)
		}
	}

	c.JSON(200, models.Channels{
		Channels: allChannels,
	})
}

func GetChannel(c *gin.Context) {
	publisherToken := c.Param("publisherToken")

	if publisherChannel, ok := publisherChannels[publisherToken]; ok {
		c.JSON(200, publisherChannel)
	} else {
		c.JSON(404, models.Response{
			Message: "Publisher token not found",
			Type:    models.ResponseErr,
		})
	}
}

func EditChannel(c *gin.Context) {
	publisherToken := c.Param("publisherToken")

	var newChannelData models.Channel
	err := c.BindJSON(&newChannelData)

	if err != nil {
		c.JSON(500, models.Response{
			Type:    models.ResponseErr,
			Message: "Invalid json passed",
		})
		return
	}

	if publisherChannel, ok := publisherChannels[publisherToken]; ok {
		publisherChannel.Tag = newChannelData.Tag
		publisherChannel.SubscriberMessageHookUrl = newChannelData.SubscriberMessageHookUrl
		publisherChannels[publisherToken] = publisherChannel

		c.JSON(200, publisherChannel)
	} else {
		c.JSON(500, models.Response{
			Type:    models.ResponseErr,
			Message: "Publisher token not found",
		})
	}
}
