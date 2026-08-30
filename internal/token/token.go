package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string, ttl time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), ttl: ttl}
}

type Claims struct {
	Document string `json:"document"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (i *Issuer) Issue(subject, document, name, role string, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(i.ttl)

	claims := Claims{
		Document: document,
		Name:     name,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    "postech-tc3-lambda-auth",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}
