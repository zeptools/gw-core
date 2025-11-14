package svc

type Service interface {
	Start() error // bootstrapping error only
	Stop()
	// Done - shutdown error channel
	// Since consumed by framework.Core only, Do Not Close the channel in a method
	Done() <-chan error
	Name() string
}
