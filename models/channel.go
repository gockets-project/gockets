package models

const ChannelCloseSignal  = 2
const ChannelSignalOk = 0
const ChannelSignalError = 1

type Channel struct {
	PublisherToken    			string      `json:"publisher_token"`
	SubscriberToken   			string		`json:"subscriber_token"`
	SubscriberMessageHookUrl	string		`json:"subscriber_message_hook_url"`
	Listeners					int			`json:"listeners"`
	
	SubscriberChannel 			chan string `json:"-"`
	SubscriberMessagesChannel	chan string `json:"-"`
	PublisherChannel  			chan int	`json:"-"`
	ResponseChannel				chan int	`json:"-"`
}
