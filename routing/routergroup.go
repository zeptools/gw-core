package routing

import (
	"log"
	"net/http"
	"strings"
)

type RouteGroup struct {
	Router          // [Embedded Interface]
	Prefix          string
	HandlerWrappers []HandlerWrapper // Group Handler Wrappers
}

// Ensure RouteGroup[any] implements Router
var _ Router = (*RouteGroup)(nil)

// Handle registers a route pattern
func (g *RouteGroup) Handle(subpattern string, handler http.Handler, handlerWrappers ...HandlerWrapper) {
	var (
		subPatternParts []string
		subpath         string
		method          string
		fullPattern     string
	)

	subPatternParts = strings.SplitN(subpattern, " ", 2)
	if len(subPatternParts) == 2 {
		// subpattern "<method> <subpath>" -> fullpattern "<method> <groupPrefix><subpath>"
		// method: e.g. GET, POST
		method = subPatternParts[0]
		subpath = subPatternParts[1]
		fullPattern = method + " " + g.Prefix + subpath
	} else {
		fullPattern = g.Prefix + subpattern
	}

	if strings.Contains(fullPattern, "//") {
		log.Fatalf("[ERROR] Can't Register Router Pattern %s", fullPattern)
	}

	// Wrapping the Handler (Nesting) by the HandlerWrappers into the Actual Handler
	// Wrapped Handler = grpHndWrapr1 (
	//						<Group-PreAction1>
	//						...
	//						grpHndWraprN (
	//							<Group-PreActionN>
	//							hndWrapr1 (
	//								<Individual-PreAction1>
	//								...
	//								hndWraprN (
	//									<Individual-PreActionN>
	//									handler
	//									<Individual-PostActionN>
	//								)
	//								...
	//								<Individual-PostAction1>
	// 							)
	//							<Group-PostActionN>
	//						)
	//						...
	//						<Group-PostAction1>
	//					)
	// 1. Pre-action order:
	//		grpHndWrapr1 -> ... -> grpHndWraprN -> hndWrapr1 -> ... -> hndWraprN
	// 2. handler.ServeHTTP(w,r)
	// 3. Post-action order:
	//		grpHndWrapr1 <- ... <- grpHndWraprN <- hndWrapr1 <- ... <- hndWraprN
	wrappedHandler := handler
	for i := len(handlerWrappers) - 1; i >= 0; i-- {
		wrappedHandler = handlerWrappers[i].Wrap(wrappedHandler)
	}
	for i := len(g.HandlerWrappers) - 1; i >= 0; i-- {
		wrappedHandler = g.HandlerWrappers[i].Wrap(wrappedHandler)
	}
	// Register the fullPattern with the WrappedHandler
	g.Router.Handle(fullPattern, wrappedHandler)
}

func (g *RouteGroup) HandleFunc(subpattern string, handleFunc func(http.ResponseWriter, *http.Request), handlerWrappers ...HandlerWrapper) {
	g.Handle(subpattern, http.HandlerFunc(handleFunc), handlerWrappers...)
}

// Group on *RouteGroup makes a Subgroup
//
//	router.Group("/foo/", func(foo *RouteGroup) {        // RouteGroup for "/foo/..."
//	  foo.Handle("GET bar", foobarGetHandler)            // "GET /foo/bar"
//
//	  foo.Group("baz/", func(foobaz *RouteGroup) {		 // RouteGroup for "/foo/baz/..." = Subgroup of "/foo/"
//	    foobaz.Handle("GET baas", foobazbaasGetHandler)  // "GET /foo/baz/baas"
//	    foobaz.Handle("POST bam", foobazbamPostHandler)  // "POST /foo/baz/bam"
//	  }
//	}
func (g *RouteGroup) Group(subPrefix string, batch func(*RouteGroup), handlerWrappers ...HandlerWrapper) *RouteGroup {
	subg := &RouteGroup{
		Router:          g.Router,                                      // same router
		Prefix:          g.Prefix + subPrefix,                          // extended prefix
		HandlerWrappers: append(g.HandlerWrappers, handlerWrappers...), // handlerwrappers appended
	}

	batch(subg)

	return subg // to do more with this routegroup if any
}
