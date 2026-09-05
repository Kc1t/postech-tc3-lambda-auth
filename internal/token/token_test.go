package token_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Kc1t/postech-tc3-lambda-auth/internal/token"
)

const secret = "segredo-de-teste"

func TestIssue(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	issuer := token.NewIssuer(secret, 15*time.Minute)

	subject := token.Subject{
		ID:       "req-1",
		Name:     "João Silva",
		Email:    "joao@email.com",
		Document: "52998224725",
	}

	signed, expiresAt, err := issuer.Issue(subject, now)
	if err != nil {
		t.Fatalf("esperava sucesso, veio %v", err)
	}

	if !expiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Errorf("esperava expiracao em %v, veio %v", now.Add(15*time.Minute), expiresAt)
	}

	parsed, err := jwt.Parse(signed, func(*jwt.Token) (interface{}, error) { return []byte(secret), nil })
	if err != nil || !parsed.Valid {
		t.Fatalf("token nao validou: %v", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("claims em formato inesperado")
	}

	expected := map[string]string{
		"sub":      "req-1",
		"role":     "client",
		"email":    "joao@email.com",
		"name":     "João Silva",
		"document": "52998224725",
	}

	for key, want := range expected {
		if got, _ := claims[key].(string); got != want {
			t.Errorf("claim %q: esperava %q, veio %q", key, want, got)
		}
	}
}

func TestIssue_RejectedByWrongSecret(t *testing.T) {
	issuer := token.NewIssuer(secret, time.Minute)

	signed, _, err := issuer.Issue(token.Subject{ID: "req-1"}, time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("esperava sucesso, veio %v", err)
	}

	if _, err := jwt.Parse(signed, func(*jwt.Token) (interface{}, error) { return []byte("outro"), nil }); err == nil {
		t.Error("esperava falha de assinatura com segredo diferente")
	}
}
