package sqldb

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
	NextResultSet() bool
}

type Row interface {
	Scan(dest ...any) error
}

type Result interface {
	RowsAffected() (int64, error)
	LastInsertId() (int64, error)
}
