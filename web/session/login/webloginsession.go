package login

type WebLoginSessionInfo[T comparable] struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       T      `json:"uid"`
	Key          string `json:"-"`
}
