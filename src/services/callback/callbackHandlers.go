package callback

import (
	"bytes"
	"gockets/models"
	"net/http"
)

func HandleSentData(channel *models.Channel) {
	for {
		select {
		case message := <-channel.SubscriberMessagesChannel:
			go sendDataToHook(channel, message)
			break
		}
	}
}

func sendDataToHook(channel *models.Channel, data string) {
	req, _ := http.NewRequest("POST", *channel.SubscriberMessageHookUrl, bytes.NewBuffer([]byte(data)))
	client := &http.Client{}
	_, _ = client.Do(req)
	return
}
