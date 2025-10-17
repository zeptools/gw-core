package routing

import "net/http"

// HandlerWrapper has Wrap method which acts as a middleware by wrapping an http.Handler
// prepending and appending some additinonal logic wrapping the handler's ServeHTTP(w,r)
// and then returns a new http.Handler which can wrap another or can be wrapped by another
type HandlerWrapper interface {
	Wrap(http.Handler) http.Handler
}
