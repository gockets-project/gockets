package helpers

import (
	"encoding/json"
	"net/http"
)

func WriteJsonResponse(w http.ResponseWriter, o interface{}) {
	preparedJson, _ := json.Marshal(o)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(preparedJson)
}
