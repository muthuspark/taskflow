package api

import (
	"encoding/json"
	"net/http"
)

// Response is the standard API response wrapper
type Response struct {
	Data   interface{} `json:"data,omitempty"`
	Status string      `json:"status,omitempty"`
	Error  string      `json:"error,omitempty"`
	Code   string      `json:"code,omitempty"`
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if statusCode >= 200 && statusCode < 300 {
		response := Response{
			Data:   data,
			Status: "success",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		json.NewEncoder(w).Encode(data)
	}
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, statusCode int, errorMsg, errorCode string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Error: errorMsg,
		Code:  errorCode,
	}
	json.NewEncoder(w).Encode(response)
}
