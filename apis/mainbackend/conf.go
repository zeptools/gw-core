package mainbackend

type Conf struct {
	Host                        string            `json:"host"`
	ClientID                    string            `json:"client_id"` // ID of this App as a Client of the MainBackendAPI
	ReissueAccessTokenEndpoint  string            `json:"reissue_access_token"`
	ReissueIdTokenEndpoint      string            `json:"reissue_id_token"`
	VerifyExternalCodeEndpoints map[string]string `json:"verify_external_code"`
}
