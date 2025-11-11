package conf

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zeptools/gw-core/web/session"
)

func (c *Core[B, U]) WebSessionIDToKVDBKey(sessionID string) string {
	return c.AppName + "_wsession:" + sessionID
}

func (c *Core[B, U]) FindWebSessionInKVDB(ctx context.Context, sessionID string) (bool, error) {
	return c.BackendKVDBClient.Exists(ctx, c.WebSessionIDToKVDBKey(sessionID))
}

func (c *Core[B, U]) CheckWebSessionFromCookie(ctx context.Context, r *http.Request) bool {
	webSessionCookie, err := r.Cookie(session.CookieName)
	if err != nil {
		return false
	}
	webSessionId, err := c.WebSessionManager.Conf.Cipher.DecodeDecrypt(webSessionCookie.Value) // []byte
	if err != nil {
		return false
	}
	found, err := c.FindWebSessionInKVDB(ctx, string(webSessionId))
	if err != nil {
		return false
	}
	return found
}

func (c *Core[B, U]) SetWebSessionCookie(w http.ResponseWriter, webSessionId string) error {
	encWebSessionId, err := c.WebSessionManager.Conf.Cipher.EncryptEncode([]byte(webSessionId))
	if err != nil {
		return fmt.Errorf("failed to encrypt web login session id. %v", err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:  session.CookieName,
		Value: encWebSessionId,
		Path:  "/", // Subpaths will get this cookie.
		// Domain: // Cannot be set with `__Host-`
		HttpOnly: true, // JS cannot read it
		Secure:   true, // only sent over HTTPS
		MaxAge:   c.WebSessionManager.Conf.ExpireHardcap,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (c *Core[B, U]) RemoveWebSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName,
		Path:     "/",
		MaxAge:   -1, // Delete
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
