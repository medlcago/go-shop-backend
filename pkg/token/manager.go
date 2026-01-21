package token

const (
	AccessTokenType  = "x-access"
	RefreshTokenType = "x-refresh"
)

type Manager interface {
	GenerateAccessToken(payload map[string]any) (string, error)
	GenerateRefreshToken(payload map[string]any) (string, error)
	ValidateToken(tokenString string) (map[string]interface{}, error)
}
