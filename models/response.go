package models

const ResponseErr = "ERR"
const ResponseInf = "INF"
const ResponseOK = "OK"

type Response struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
