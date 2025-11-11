package security

type AuthUserID interface {
	~int64 | ~string
}
