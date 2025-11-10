package login

import "github.com/zeptools/gw-core/security"

type WebLoginSessionConf struct {
	LoginPath     string `json:"login_path"`
	EncryptionKey string `json:"enckey"`
	ExpireSliding int    `json:"expire_sliding"`
	ExpireHardcap int    `json:"expire_hardcap"`
	MaxCntPerUser int64  `json:"max_cnt_per_user"`

	Cipher *security.XChaCha20Poly1305Cipher `json:"-"`
}
