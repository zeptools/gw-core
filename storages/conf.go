package storages

import "github.com/zeptools/gw-core/storages/keystores"

type Conf struct {
	KeyStoreConf keystores.Conf `json:"key_store"`
}
