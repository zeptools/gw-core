package responses

type Message struct {
	Type    string `json:"type"` // "error", etc
	Message string `json:"message"`
	Code    int    `json:"code"` // application-level logic code
}
