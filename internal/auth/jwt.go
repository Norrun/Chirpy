package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()),
		Subject:   userID.String(),
	})

	what, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return what, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var empty uuid.UUID
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return empty, err
	}
	subject := claims.Subject
	id, err := uuid.Parse(subject)
	if err != nil {
		return empty, err
	}
	return id, nil

}
