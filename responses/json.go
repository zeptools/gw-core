package responses

import (
	"encoding/json/v2"
	"log"
	"net/http"
)

// WriteJSONBytes Write Already Encoded JSON Bytes into the Response
// JSONBytes, err := json.Marshal(payload any)
func WriteJSONBytes(w http.ResponseWriter, HTTPStatusCode int, JSONBytes []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(HTTPStatusCode) // Response Header Sent & Frozen
	if _, err := w.Write(JSONBytes); err != nil {
		log.Printf("[ERROR] Writing JSON to Response: %v", err)
	}
}

// EncodeWriteJSON Encode & Write Payload as JSON Stream to the Response
func EncodeWriteJSON(w http.ResponseWriter, HTTPStatusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(HTTPStatusCode) // Response Header Sent & Frozen
	if err := json.MarshalWrite(w, payload); err != nil {
		log.Printf("[ERROR] failed to write JSON Stream to Response: %v", err)
	}
}

// WriteSimpleErrorJSON is a helper func same as EncodeWriteJSON
// but wrapping a string message into a simple Message without app logic code
func WriteSimpleErrorJSON(w http.ResponseWriter, HTTPStatusCode int, msg string) {
	payload := Message{Type: "error", Message: msg}
	EncodeWriteJSON(w, HTTPStatusCode, payload)
}
