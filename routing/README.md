# Routing System

## - Aliasing with Type Instantiation

```
type Router = routing.BaseRouter[Env]

type RouteGroup = routing.RouteGroup[Env]
```

where `Env` is the Concrete Type in your application.

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
	routing.BaseRouter[Env]
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ...
    ctx := foo.WithFooConf(r.Context(), bar)
    router.ServeMux.ServeHTTP(w, r.WithContext(ctx))
    ...
}
```

still, your RouteGroup can be an alias with type instantiation:

```
type RouteGroup = routing.RouteGroup[Env]
```

