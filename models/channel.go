package models

type Channel struct {
	PublisherToken    string      `json:"publisher_token"`
	SubscriberToken   string      `json:"subscriber_token"`
	SubscriberChannel chan string `json:"-"`
}
