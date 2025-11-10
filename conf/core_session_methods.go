package conf

import (
	"context"
	"net/http"

	"github.com/zeptools/gw-core/web/session"
)

func (c *Core[B]) WebSessionIDToKVDBKey(sessionID string) string {
	return c.AppName + "_wsession:" + sessionID
}

func (c *Core[B]) FindWebSessionInKVDB(ctx context.Context, sessionID string) (bool, error) {
	return c.BackendKVDBClient.Exists(ctx, c.WebSessionIDToKVDBKey(sessionID))
}

func (c *Core[B]) CheckWebSessionFromCookie(ctx context.Context, r *http.Request) bool {
	webSessionCookie, err := r.Cookie(session.CookieName)
	if err != nil {
		return false
	}
	webSessionId, err := c.WebSessionConf.Cipher.DecodeDecrypt(webSessionCookie.Value) // []byte
	if err != nil {
		return false
	}
	found, err := c.FindWebSessionInKVDB(ctx, string(webSessionId))
	if err != nil {
		return false
	}
	return found
}
