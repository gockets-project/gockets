# Gockets
**Gockets** is daemon written in Golang to give languages, like PHP a middleware for REST-oriented communication with Websockets.

## Build
1. Clone repository
2. Change current directory to repository folder
3. Execute `go build main.go`

## Running

Execute binary created after building.

## Usage
By default daemon is listening to **8844** port. Change it in `main` function if needed.

`GET` `/channel/prepare` - prepares channel and returns JSON object which contains `publisher_key` and `subscriber_key`.   
`GET` `/channel` - list of all channels awaiting publishing or subscription.  
`GET` `/channel/subscribe/{subscriber_key}` - creates a Websocket connection with channel referenced by `subscriber_key`.  
`POST` `/channel/publish/{publisher_key}` - pushes data to a Websocket connection passed in `data` argument of request.

## Use case

Create a channel and share a subscriber key to any Front-End application you want to communicate via Websocket. After doing so, with any kind of mechanisms, like Events in Laravel Framework hook pushing data into Websocket.

### TODO

* Implementation of graceful close of Websocket connection
* Implementation of 2-way communication with URL-Hook



