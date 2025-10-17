package keystores

type Conf struct {
	Type          string `json:"type"`
	PrivateKeyDir string `json:"private_key_dir"`
	PublicKeyDir  string `json:"public_key_dir"`
}
