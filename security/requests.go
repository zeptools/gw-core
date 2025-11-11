package security

type AuthRequestBody struct {
	AuthClientID string `json:"auth_client_id"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"` // Required for Security Check. RFC 6749 §4.1.3
}

type ReissueAccessTokenRequestBody struct {
	RefreshToken string `json:"refresh_token"`
	UserIDStr    string `json:"uid"` // Optional for Double-Checking
}

func ExtractBearerToken(header string) string {
	const prefix = "Bearer "
	prefixLen := len(prefix)
	if len(header) > prefixLen && header[:prefixLen] == prefix {
		return header[prefixLen:]
	}
	return ""
}
