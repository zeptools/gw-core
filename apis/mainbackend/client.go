package mainbackend

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
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

func (c *Client) JWKSFileResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	upstrRes, err := c.RequestJWKS(ctx) // *http.Response
	if err != nil {
		responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	if upstrRes.StatusCode == http.StatusNotFound {
		// 404 not found -> raw error message sent before wrapped into JSON
		responses.WriteSimpleErrorJSON(w, http.StatusNotFound, fmt.Sprintf("%v", responses.HTTPErrorNotFound))
		return
	}
	defer func() {
		if closeErr := upstrRes.Body.Close(); closeErr != nil {
			log.Printf("[WARN] %v", closeErr)
		}
	}()
	w.Header().Set("Content-Type", "application/jwk-set+json")
	w.WriteHeader(upstrRes.StatusCode)
	_, err = io.Copy(w, upstrRes.Body)
	if err != nil {
		responses.WriteSimpleErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
}

// RequestJSON sends a request and returns the response.
// The caller is responsible for closing response.Body.
func (c *Client) RequestJSON(ctx context.Context, accessToken string, method string, endpoint string) (*http.Response, error) {
	upstrUrl := c.Conf.Host + endpoint
	upstrReq, err := http.NewRequestWithContext(ctx, method, upstrUrl, nil)
	if err != nil {
		return nil, err
	}

	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Authorization", "Bearer "+accessToken)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/json")

	upstrRes, err := c.Do(upstrReq)
	if err != nil {
		return nil, err
	}
	return upstrRes, nil
}

// RequestReissueAccessTokenWithRefreshToken requests the MainBackend to reissue access token only with refresh token
func (c *Client) RequestReissueAccessTokenWithRefreshToken(ctx context.Context, refreshToken string) (*http.Response, error) {
	upstrURL := c.Conf.Host + c.Conf.ReissueAccessTokenEndpoint
	upstrReqBody := security.ReissueAccessTokenRequestBody{
		RefreshToken: refreshToken,
	}
	upstrReqBodyBytes, err := json.Marshal(upstrReqBody)
	if err != nil {
		return nil, err
	}
	upstrReq, err := http.NewRequestWithContext(ctx, http.MethodPost, upstrURL, bytes.NewReader(upstrReqBodyBytes))
	if err != nil {
		return nil, err
	}
	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/json")
	return c.Do(upstrReq)
}

// RequestReissueAccessTokenWithRefreshTokenAndUserID requests the MainBackend to reissue access token with refresh token and user id
func (c *Client) RequestReissueAccessTokenWithRefreshTokenAndUserID(ctx context.Context, refreshToken string, userIDStr string) (*http.Response, error) {
	upstrURL := c.Conf.Host + c.Conf.ReissueAccessTokenEndpoint
	upstrReqBody := security.ReissueAccessTokenRequestBody{
		RefreshToken: refreshToken,
		UserIDStr:    userIDStr,
	}
	upstrReqBodyBytes, err := json.Marshal(upstrReqBody)
	if err != nil {
		return nil, err
	}
	upstrReq, err := http.NewRequestWithContext(ctx, http.MethodPost, upstrURL, bytes.NewReader(upstrReqBodyBytes))
	if err != nil {
		return nil, err
	}
	upstrReq.Header.Set("Client-Id", c.Conf.ClientID)
	upstrReq.Header.Set("Content-Type", "application/json")
	upstrReq.Header.Set("Accept", "application/json")
	return c.Do(upstrReq)
}
