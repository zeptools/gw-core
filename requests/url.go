package requests

import (
	"fmt"
	"net/http"
)

func FullURL(req *http.Request) string {
	scheme := ""
	if req.TLS != nil {
		scheme = "https"
	} else {
		scheme = req.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "http"
		}
	}
	return fmt.Sprintf("%s://%s%s", scheme, req.Host, req.URL.RequestURI())
}
