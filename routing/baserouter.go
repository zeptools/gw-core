package routing

import "net/http"

type BaseRouter[T any] struct {
	*http.ServeMux // Embedded
	Env            *T
}

// Ensure BaseRouter[any] implements Router
var _ Router = (*BaseRouter[any])(nil)

// Handle registers a route pattern
func (r *BaseRouter[T]) Handle(pattern string, handler http.Handler, handlerWrappers ...HandlerWrapper) {
	wrappedHandler := handler
	for i := len(handlerWrappers) - 1; i >= 0; i-- {
		wrappedHandler = handlerWrappers[i].Wrap(wrappedHandler)
	}
	r.ServeMux.Handle(pattern, wrappedHandler)
}

func (r *BaseRouter[T]) HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request), handlerWrappers ...HandlerWrapper) {
	r.Handle(pattern, http.HandlerFunc(handleFunc), handlerWrappers...)
}

// Group lets you register routes under a common Prefix + middleware.
func (r *BaseRouter[T]) Group(prefix string, batch func(*RouteGroup[T]), handlerWrappers ...HandlerWrapper) *RouteGroup[T] {
	rg := &RouteGroup[T]{
		Router:          r,
		Env:             r.Env,
		Prefix:          prefix,
		HandlerWrappers: handlerWrappers,
	}

	batch(rg)

	return rg // to do more with this routegroup if any
}
