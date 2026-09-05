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
