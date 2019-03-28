package helpers

import (
	"encoding/json"
	"gockets/src/services/logger"
	"net/http"
)

func WriteJsonResponse(w http.ResponseWriter, o interface{}) {
	preparedJson, _ := json.Marshal(o)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
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
