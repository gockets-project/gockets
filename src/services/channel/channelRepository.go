package channel

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"gockets/models"
	"gockets/src/services/logger"
	"net/http"
	"strconv"
	"time"
)

var PublisherChannels = make(map[string]*models.Channel)

var SubscriberChannels = make(map[string]*models.Channel)

func PrepareChannel(w http.ResponseWriter, r *http.Request) {
	ll.Log.Debugf("Channel prepared by: %s", r.Host)
	var channel models.Channel
	for {
		channel = generateChannel(r)
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

func GetChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var preparedJson []byte
	w.Header().Set("Content-Type", "application/json")
	if publisherChannel, ok := PublisherChannels[vars["publisherToken"]]; ok {
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
