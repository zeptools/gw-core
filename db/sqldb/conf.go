package sqldb

type Conf struct {
	Type string `json:"type"` // mysql, pgsql, mssql, oracle, maria, sqlite, ...
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	PW   string `json:"pw"`
	DB   string `json:"db"`
	TZ   string `json:"tz"`  // Connection Timezone
	DSN  string `json:"dsn"` // To Overwrite Default DSN
}
