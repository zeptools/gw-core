package session

import "github.com/zeptools/gw-core/security"

type Manager struct {
	Conf    Conf
	AppName string // for session key, etc.
	Cipher  *security.XChaCha20Poly1305Cipher
}

func (m *Manager) WebSessionIDToKVDBKey(sessionID string) string {
	return m.AppName + "_wsession:" + sessionID
}
