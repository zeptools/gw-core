package kvdb

type Conf struct {
	Type string `json:"type"`
	Host string `json:"host"`
	Port int    `json:"port"`
	//Driver string `json:"driver"`
	PW string `json:"pw"`
	DB int    `json:"db"` // optional db number e.g. redis
}
