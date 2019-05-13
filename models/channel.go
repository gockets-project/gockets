package models

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

const ChannelCloseSignal = 2
const ChannelSignalOk = 0
const ChannelSignalError = 1

type Channel struct {
	PublisherToken           string  `json:"publisher_token"`
	SubscriberToken          string  `json:"subscriber_token"`
	SubscriberMessageHookUrl *string `json:"subscriber_message_hook_url"`
	Tag                      *string `json:"tag"`
	Listeners                int     `json:"listeners"`

	SubscriberChannel         chan string `json:"-"`
	SubscriberMessagesChannel chan string `json:"-"`
	PublisherChannel          chan int    `json:"-"`
	ResponseChannel           chan int    `json:"-"`
}

func GenerateChannel(r *http.Request) Channel {

	decoder := json.NewDecoder(r.Body)
	var c Channel
	_ = decoder.Decode(&c)

	hasher := md5.New()
	timeString := strconv.FormatInt(time.Now().Unix(), 10)
	hasher.Write([]byte(timeString))
	publisherKey := hex.EncodeToString(hasher.Sum(nil))

	hasher.Write([]byte(publisherKey))
	subscriberKey := hex.EncodeToString(hasher.Sum(nil))

	return Channel{
		PublisherToken:           publisherKey,
		SubscriberToken:          subscriberKey,
		SubscriberMessageHookUrl: c.SubscriberMessageHookUrl,
		Listeners:                0,
		Tag:                      c.Tag,

		ResponseChannel:           make(chan int),
		PublisherChannel:          make(chan int),
		SubscriberChannel:         make(chan string),
		SubscriberMessagesChannel: make(chan string),
	}
}
