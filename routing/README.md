# Routing System

## - Just Using Bundled BaseRouter and RouteGroup

```
type Router = routing.BaseRouter
type RouteGroup = routing.RouteGroup
```

## - Base Middleware to Apply Entire Requests
```
server := &http.Server{
    Addr:    env.Listen,
    Handler: HttpHandlerWrapper(router),
}
```
where `HttpHandlerWrapper` is a `func(http.Handler) http.Handler`

## - Overriding Methods by Embedding
You can write your own router.

e.g.
```
type Router struct {
	routing.BaseRouter
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ...
    ctx := foo.WithFooConf(r.Context(), bar)
    router.ServeMux.ServeHTTP(w, r.WithContext(ctx))
    ...
}

// still, your RouteGroup can be an alias with type instantiation:
type RouteGroup = routing.RouteGroup
```
