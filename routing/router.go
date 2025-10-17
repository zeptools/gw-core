package routing

import "net/http"

type Router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Handle(pattern string, handler http.Handler, handlerWrappers ...HandlerWrapper)
	HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request), handlerWrappers ...HandlerWrapper)
}
