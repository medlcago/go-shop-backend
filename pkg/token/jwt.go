package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWT struct {
	secretKey               string
	issuer                  string
	accessTokenExpiredTime  time.Duration
	refreshTokenExpiredTime time.Duration
	partialTokenExpiredTime time.Duration
}

func NewJWT(
	secretKey string,
	issuer string,
	accessTokenExpiredTime time.Duration,
	refreshTokenExpiredTime time.Duration,
	partialTokenExpiredTime time.Duration,
) *JWT {
	return &JWT{
		secretKey:               secretKey,
		issuer:                  issuer,
		accessTokenExpiredTime:  accessTokenExpiredTime,
		refreshTokenExpiredTime: refreshTokenExpiredTime,
		partialTokenExpiredTime: partialTokenExpiredTime,
	}
}

func (j JWT) GenerateAccessToken(payload Payload) (string, *UserClaims, error) {
	exp := time.Now().UTC().Add(j.accessTokenExpiredTime)
	return j.generateToken(payload, AccessTokenType, exp)
}

func (j JWT) GenerateRefreshToken(payload Payload) (string, *UserClaims, error) {
	exp := time.Now().UTC().Add(j.refreshTokenExpiredTime)
	return j.generateToken(payload, RefreshTokenType, exp)
}

func (j JWT) GeneratePartialToken(payload Payload) (string, *UserClaims, error) {
	exp := time.Now().UTC().Add(j.partialTokenExpiredTime)
	return j.generateToken(payload, PartialTokenType, exp)
}

func (j JWT) ValidateToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(j.issuer),
	)

	if err != nil {
		return nil, &ErrInvalidToken{
			Err: err,
		}
	}

	if !token.Valid {
		return nil, &ErrInvalidToken{
			Err: jwt.ErrInvalidKey,
		}
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, &ErrInvalidToken{
			Err: jwt.ErrTokenInvalidClaims,
		}
	}

	return claims, nil
}

func (j JWT) generateToken(payload Payload, tokenType string, exp time.Time) (string, *UserClaims, error) {
	claims := &UserClaims{
		Payload:   payload,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        uuid.NewString(),
			Subject:   payload.UserID,
			Issuer:    j.issuer,
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", nil, &ErrInvalidToken{
			Err: err,
		}
	}

	return token, claims, nil
}
