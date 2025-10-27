package web

import (
	"context"
	"net/http"

	"github.com/zeptools/gw-core/service"
)

type Service struct {
	Ctx    context.Context    // Service Context
	Cancel context.CancelFunc // Service Context CancelFunc
	state  int                // internal service state
	Server *http.Server
}

func NewService(parentCtx context.Context, addr string, router http.Handler) *Service {
	svcCtx, svcCancel := context.WithCancel(parentCtx)
	return &Service{
		Ctx:    svcCtx,
		Cancel: svcCancel,
		state:  service.StateREADY,
		Server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func (s *Service) Start() {

}
