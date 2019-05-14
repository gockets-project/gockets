package helpers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
)

func WriteJsonResponse(w http.ResponseWriter, o interface{}, respCode int) {
	preparedJson, _ := json.Marshal(o)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respCode)
	w.Write(preparedJson)
}

func LookupName(hostname string) (string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return "", err
	}

	if len(ips) > 0 {
		return ips[0].String(), nil
	}

	return "", errors.New("no ips defined")
}

func SliceContains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
