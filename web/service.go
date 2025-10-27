package web

import "net/http"

type Service struct {
	Server *http.Server
}

func NewService(addr string, router http.Handler) *Service {
	return &Service{
		Server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}
