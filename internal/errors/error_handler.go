package errors

import (
	"net/http"
)

func HandleError(status int, message string, w http.ResponseWriter) {
	w.WriteHeader(status)
	w.Write([]byte(message))
}
