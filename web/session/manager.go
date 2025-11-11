package session

import "github.com/zeptools/gw-core/security"

type Manager struct {
	Conf Conf

	Cipher *security.XChaCha20Poly1305Cipher
}
