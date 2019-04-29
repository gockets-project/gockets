package helpers

import (
	"encoding/json"
	"errors"
	"gockets/src/services/logger"
	"net"
	"net/http"
)

func WriteJsonResponse(w http.ResponseWriter, o interface{}, respCode int) {
	preparedJson, _ := json.Marshal(o)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respCode)
	w.Write(preparedJson)
}

func LogError(e error) {
	if e != nil {
		ll.Log.Error(e)
	}
}

func LogErrorf(f string, e error) {
	if e != nil {
		ll.Log.Errorf(f, e)
	}
}

func LookupName(hostname string) (string, error) {
	ips, err := net.LookupIP("127.0.0.1")
	if err != nil {
		return "", err
	}

	if len(ips) > 0 {
		return ips[0].String(), nil
	}

	return "", errors.New("no ips defined")
}
