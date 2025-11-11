package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/zeptools/gw-core/db/kvdb"
	"github.com/zeptools/gw-core/security"
)

type Manager struct {
	Conf              Conf
	Cipher            *security.XChaCha20Poly1305Cipher
	AppName           string // for session key, etc.
	SessionLocks      *sync.Map
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

// CreateWebLoginSession creates a Session in Key-value Database and Returns its Session ID
func (m *Manager) CreateWebLoginSession(ctx context.Context, accessToken string, refreshToken string, uid int64) (string, error) {
	webSessionID, err := GenerateWebSessionID()
	if err != nil {
		return "", err
	}
	// Store session_id in KvDB with access_token and refresh_token
	key := m.WebSessionIDToKVDBKey(webSessionID)
	if err = m.BackendKVDBClient.SetFields(ctx, key, map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"uid":           uid,
	}); err != nil {
		return "", err
	}

	slidingExpiration := time.Duration(m.Conf.ExpireSliding) * time.Second
	hardcapExpiration := time.Duration(m.Conf.ExpireHardcap) * time.Second

	ok, err := m.BackendKVDBClient.Expire(ctx, key, slidingExpiration)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("failed to set session expiration")
	}

	if m.Conf.MaxCntPerUser > 0 {
		usrSessionListKey := fmt.Sprintf("%s_wsessions:%d", m.AppName, uid)
		// SessionList Lock (User Level Lock)
		mu, _ := m.SessionLocks.LoadOrStore(usrSessionListKey, &sync.Mutex{})
		mutex := mu.(*sync.Mutex)

		mutex.Lock() // waits until this gets the lock if it's locked by another goroutine
		defer mutex.Unlock()
		// No need to delete the lock. Keep it for reusing it's tiny in memory.
		// By keeping it, no overhead to create(LoadOrStore)/delete the lock
		// If you wants to delete the lock every time:
		//defer func() {
		//	mutex.Unlock()
		//	env.SessionLocks.Delete(usrSessionListKey)
		//}()

		if err = m.BackendKVDBClient.Push(ctx, usrSessionListKey, webSessionID); err != nil {
			return "", err
		}
		sessionCnt, err := m.BackendKVDBClient.Len(ctx, usrSessionListKey)
		if err != nil {
			return "", err
		}
		if sessionCnt > m.Conf.MaxCntPerUser {
			diff := sessionCnt - m.Conf.MaxCntPerUser
			sessionsToDel, err := m.BackendKVDBClient.Range(ctx, usrSessionListKey, 0, diff-1) // []string
			if err != nil {
				return "", err
			}
			var keysToDel []string
			for _, v := range sessionsToDel {
				keysToDel = append(keysToDel, m.WebSessionIDToKVDBKey(v))
			}
			_, _ = m.BackendKVDBClient.Delete(ctx, keysToDel...)
			if err = m.BackendKVDBClient.Trim(ctx, usrSessionListKey, diff, -1); err != nil {
				return "", err
			}
			_, _ = m.BackendKVDBClient.Expire(ctx, usrSessionListKey, hardcapExpiration)
		}
	}

	return webSessionID, nil
}
