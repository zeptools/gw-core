package responses

import (
	"bytes"
	jsonv1 "encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/zeptools/gw-core/requests"
)

type EchoHandler struct {
	MaxMemoryMB int64
}

// ToDo: Use json/v2

func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resPayload := map[string]any{
		"url":    requests.FullURL(r),
		"method": r.Method,
		"header": r.Header,
	}

	if !requests.HasBody(r) {
		EncodeWriteJSON(w, http.StatusOK, resPayload)
		return
	}

	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			log.Printf("[ERROR] %v", closeErr)
		}
	}()

	rBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("Failed to Read Data: %v", err))
		return
	}

	rBodyPayload := make(map[string]interface{})
	rBodyPayload["raw"] = string(rBodyBytes)

	// Since we already consumed r.Data with io.ReadAll(r.Data),
	// Reassign r.Data to a No-op closer Reader on a copied buffer like rewinding r.Data
	r.Body = io.NopCloser(bytes.NewReader(rBodyBytes))

	rContentType := r.Header.Get("Content-Type")

	switch true {

	case strings.HasPrefix(rContentType, "application/json"):
		if jsonv1.Valid(rBodyBytes) {
			rBodyPayload["json"] = jsonv1.RawMessage(rBodyBytes)
		}

	case strings.HasPrefix(rContentType, "application/x-www-form-urlencoded"):
		if err = r.ParseForm(); err == nil {
			rBodyPayload["form"] = r.PostForm
		} else {
			rBodyPayload["form_error"] = err.Error()
		}

	case strings.HasPrefix(rContentType, "multipart/form-data"):
		if err = r.ParseMultipartForm(h.MaxMemoryMB << 20); err == nil {
			rBodyPayload["form"] = r.PostForm
			rBodyPayload["files"] = r.MultipartForm.File
		} else {
			rBodyPayload["form_error"] = err.Error()
		}

	}
	resPayload["body"] = rBodyPayload
	EncodeWriteJSON(w, http.StatusOK, resPayload)
}
