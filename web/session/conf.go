package session

import "github.com/zeptools/gw-core/security"

type Conf struct {
	EncryptionKey string                            `json:"enckey"`
	Cipher        *security.XChaCha20Poly1305Cipher `json:"-"`

	ExpireSliding int   `json:"expire_sliding"`
	ExpireHardcap int   `json:"expire_hardcap"`
	MaxCntPerUser int64 `json:"max_cnt_per_user"`

	// For Web Login Sessions
	LoginPath string `json:"login_path"`
}
