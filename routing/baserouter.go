package routing

import "net/http"

type BaseRouter struct {
	*http.ServeMux // Embedded
}

// Ensure BaseRouter[any] implements Router
var _ Router = (*BaseRouter)(nil)

// Handle registers a route pattern
func (r *BaseRouter) Handle(pattern string, handler http.Handler, handlerWrappers ...HandlerWrapper) {
	wrappedHandler := handler
	for i := len(handlerWrappers) - 1; i >= 0; i-- {
		wrappedHandler = handlerWrappers[i].Wrap(wrappedHandler)
	}
	r.ServeMux.Handle(pattern, wrappedHandler)
}

func (r *BaseRouter) HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request), handlerWrappers ...HandlerWrapper) {
	r.Handle(pattern, http.HandlerFunc(handleFunc), handlerWrappers...)
}

// Group lets you register routes under a common Prefix + middleware.
func (r *BaseRouter) Group(prefix string, batch func(*RouteGroup), handlerWrappers ...HandlerWrapper) *RouteGroup {
	g := &RouteGroup{
		Router:          r,
		Prefix:          prefix,
		HandlerWrappers: handlerWrappers,
	}

	batch(g)

	return g // to do more with this routegroup if any
}
