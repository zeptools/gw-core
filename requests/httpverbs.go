package requests

import "net/http"

func HasBody(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		// bodiless
		return false
	default:
		return true
	}
}
