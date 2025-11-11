package login

type WebLoginSessionInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserIDStr    string `json:"uid"`
	Key          string `json:"-"`
}
