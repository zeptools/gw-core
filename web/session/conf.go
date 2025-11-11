package session

type Conf struct {
	EncryptionKey string `json:"enckey"`
	ExpireSliding int    `json:"expire_sliding"`
	ExpireHardcap int    `json:"expire_hardcap"`

	// For Web Login Sessions
	LoginPath     string `json:"login_path"`
	MaxCntPerUser int64  `json:"max_cnt_per_user"`
}
