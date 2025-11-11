package session

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zeptools/gw-core/db/kvdb"
	"github.com/zeptools/gw-core/security"
)

type Manager struct {
	Conf              Conf
	Cipher            *security.XChaCha20Poly1305Cipher
	AppName           string // for session key, etc.
	BackendKVDBClient kvdb.Client
}

func (m *Manager) WebSessionIDToKVDBKey(sessionID string) string {
	return m.AppName + "_wsession:" + sessionID
}

func (m *Manager) FindWebSessionInKVDB(ctx context.Context, sessionID string) (bool, error) {
	return m.BackendKVDBClient.Exists(ctx, m.WebSessionIDToKVDBKey(sessionID))
}

func (m *Manager) CheckWebSessionFromCookie(ctx context.Context, r *http.Request) bool {
	webSessionCookie, err := r.Cookie(CookieName)
	if err != nil {
		return false
	}
	webSessionId, err := m.Cipher.DecodeDecrypt(webSessionCookie.Value) // []byte
	if err != nil {
		return false
	}
	found, err := m.FindWebSessionInKVDB(ctx, string(webSessionId))
	if err != nil {
		return false
	}
	return found
}

func (m *Manager) SetWebSessionCookie(w http.ResponseWriter, webSessionId string) error {
	encWebSessionId, err := m.Cipher.EncryptEncode([]byte(webSessionId))
	if err != nil {
		return fmt.Errorf("failed to encrypt web login session id. %v", err)
	}
	http.SetCookie(w, &http.Cookie{
		Name:  CookieName,
		Value: encWebSessionId,
		Path:  "/", // Subpaths will get this cookie.
		// Domain: // Cannot be set with `__Host-`
		HttpOnly: true, // JS cannot read it
		Secure:   true, // only sent over HTTPS
		MaxAge:   m.Conf.ExpireHardcap,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *Manager) RemoveWebSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		MaxAge:   -1, // Delete
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
