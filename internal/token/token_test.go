package token_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Kc1t/postech-tc3-lambda-auth/internal/token"
)

const secret = "segredo-de-teste"

func TestIssue(t *testing.T) {
	now := time.Now()
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

	signed, _, err := issuer.Issue(token.Subject{ID: "req-1"}, time.Now())
	if err != nil {
		t.Fatalf("esperava sucesso, veio %v", err)
	}

	if _, err := jwt.Parse(signed, func(*jwt.Token) (interface{}, error) { return []byte("outro"), nil }); err == nil {
		t.Error("esperava falha de assinatura com segredo diferente")
	}
}

func TestVerify(t *testing.T) {
	now := time.Now()
	issuer := token.NewIssuer(secret, 15*time.Minute)

	subject := token.Subject{ID: "req-1", Name: "João", Email: "j@j.com", Document: "52998224725"}
	signed, _, err := issuer.Issue(subject, now)
	if err != nil {
		t.Fatalf("esperava sucesso, veio %v", err)
	}

	claims, err := issuer.Verify(signed)
	if err != nil {
		t.Fatalf("esperava token valido, veio %v", err)
	}

	if claims.Subject != "req-1" || claims.Role != "client" || claims.Document != "52998224725" {
		t.Errorf("claims inesperadas: %+v", claims)
	}
}

func TestVerify_Rejects(t *testing.T) {
	now := time.Now()
	issuer := token.NewIssuer(secret, 15*time.Minute)

	signed, _, err := issuer.Issue(token.Subject{ID: "req-1"}, now)
	if err != nil {
		t.Fatalf("esperava sucesso, veio %v", err)
	}

	cases := map[string]struct {
		issuer *token.Issuer
		token  string
	}{
		"segredo diferente": {token.NewIssuer("outro-segredo", time.Minute), signed},
		"token vazio":       {issuer, ""},
		"token corrompido":  {issuer, signed + "x"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := tc.issuer.Verify(tc.token); err == nil {
				t.Error("esperava erro, veio nil")
			}
		})
	}
}

func TestVerify_RejectsExpired(t *testing.T) {
	issuer := token.NewIssuer(secret, time.Minute)

	signed, _, err := issuer.Issue(token.Subject{ID: "req-1"}, time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("esperava sucesso, veio %v", err)
	}

	if _, err := issuer.Verify(signed); err == nil {
		t.Error("esperava rejeicao de token expirado")
	}
}

func TestVerify_RejectsNoneAlgorithm(t *testing.T) {
	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub":  "req-1",
		"role": "client",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("falha ao montar token sem assinatura: %v", err)
	}

	if _, err := token.NewIssuer(secret, time.Minute).Verify(unsigned); err == nil {
		t.Error("esperava rejeicao do algoritmo none")
	}
}
