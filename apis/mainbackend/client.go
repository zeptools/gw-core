package mainbackend

import "net/http"

type Client struct {
	*http.Client // [Embedded]
	Conf         *Conf
}
