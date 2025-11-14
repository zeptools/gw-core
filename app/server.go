package app

type Server[B comparable] interface {
	AppCore() *Core[B]
}
