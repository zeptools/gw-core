package responses

import (
	"fmt"
	"log"
	"net/http"
)

func WritePDFBytesWithFilename(w http.ResponseWriter, filename string, PDFBytes []byte) {
	WritePDFResponseHeaders(w, filename)
	_, err := w.Write(PDFBytes)
	if err != nil {
		log.Printf("[ERROR] writing PDF to response: %v", err)
	}
}

// WritePDFResponseHeaders write HTTP response headers for PDF response. i.e. headers are frozen
func WritePDFResponseHeaders(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	w.WriteHeader(http.StatusOK) // Response Header Sent & Frozen
}
