package sqldb

import "fmt"

// ClientFactory is a callback that constructs a Client from Conf.
// It is registered with RegisterFactory and called by sqldb.New.
type ClientFactory func(conf *Conf) (Client, error)

var registry = map[string]ClientFactory{}

func RegisterFactory(dbType string, factory ClientFactory) {
	registry[dbType] = factory
}

func New(dbType string, conf *Conf) (Client, error) {
	factory, ok := registry[dbType]
	if !ok {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	return factory(conf)
}
