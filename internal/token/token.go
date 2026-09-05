package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const roleClient = "client"

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string, ttl time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), ttl: ttl}
}

type Subject struct {
	ID       string
	Name     string
	Email    string
	Document string
}

func (i *Issuer) Issue(subject Subject, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(i.ttl)

	claims := jwt.MapClaims{
		"sub":      subject.ID,
		"email":    subject.Email,
		"role":     roleClient,
		"name":     subject.Name,
		"document": subject.Document,
		"iat":      now.Unix(),
		"exp":      expiresAt.Unix(),
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

type Claims struct {
	Subject  string
	Role     string
	Email    string
	Name     string
	Document string
}

func (i *Issuer) Verify(signed string) (Claims, error) {
	parsed, err := jwt.Parse(signed, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return i.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return Claims{}, err
	}

	raw, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return Claims{}, jwt.ErrTokenInvalidClaims
	}

	claims := Claims{
		Subject:  stringClaim(raw, "sub"),
		Role:     stringClaim(raw, "role"),
		Email:    stringClaim(raw, "email"),
		Name:     stringClaim(raw, "name"),
		Document: stringClaim(raw, "document"),
	}

	if claims.Subject == "" || claims.Role == "" {
		return Claims{}, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func stringClaim(claims jwt.MapClaims, key string) string {
	value, _ := claims[key].(string)
	return value
}
