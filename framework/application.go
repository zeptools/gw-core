package framework

type Application[B comparable] interface {
	AppCore() *Core[B]
}
