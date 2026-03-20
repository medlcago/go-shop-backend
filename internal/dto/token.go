package dto

type TokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	PartialToken string `json:"partial_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
}
