package mainbackend

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"log"
	"net/http"

	"github.com/zeptools/gw-core/responses"
	"github.com/zeptools/gw-core/security"
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

// GetJWKS fetches JWKS from the Main API's .well-known URL
func (c *Client) GetJWKS(ctx context.Context) (*security.JWKS, error) {
	upstrRes, err := c.RequestJWKS(ctx)
	if err != nil {
		return nil, err
	}
	if upstrRes.StatusCode == http.StatusNotFound {
		return nil, responses.HTTPErrorNotFound
	}
	if upstrRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Status Code: %d", upstrRes.StatusCode)
	}
	defer func() {
		if err = upstrRes.Body.Close(); err != nil {
			log.Printf("[WARN] %v", err)
		}
	}()
	var jwks security.JWKS
	if err = json.UnmarshalRead(upstrRes.Body, &jwks); err != nil {
		return nil, err
	}
	return &jwks, nil
}
