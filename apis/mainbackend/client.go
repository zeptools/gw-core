package mainbackend

import (
	"context"
	"net/http"
)

type Client struct {
	*http.Client // [Embedded]
	Conf         *Conf
}

func (c *Client) RequestJWKS(ctx context.Context) (*http.Response, error) {
	upstrUrl := c.Conf.Host + "/.well-known/jwks.json"
	upstrReq, err := http.NewRequestWithContext(ctx, http.MethodGet, upstrUrl, nil) // *http.Request
	if err != nil {
		return nil, err
	}
	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/jwk-set+json")
	return http.DefaultClient.Do(upstrReq) // *http.Response
}
