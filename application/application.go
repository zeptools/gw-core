package application

type Application[B comparable] interface {
	AppCore() *Core[B]
}
