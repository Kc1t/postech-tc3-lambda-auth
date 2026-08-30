package cpf_test

import (
	"testing"

	"github.com/Kc1t/postech-tc3-lambda-auth/internal/cpf"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{"formatado", "529.982.247-25", "52998224725", true},
		{"apenas digitos", "52998224725", "52998224725", true},
		{"digito verificador errado", "52998224724", "", false},
		{"todos digitos iguais", "11111111111", "", false},
		{"tamanho invalido", "5299822472", "", false},
		{"vazio", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cpf.Validate(tc.input)

			if tc.ok && err != nil {
				t.Fatalf("esperava sucesso, veio %v", err)
			}

			if !tc.ok && err == nil {
				t.Fatal("esperava erro, veio nil")
			}

			if got != tc.want {
				t.Fatalf("esperava %q, veio %q", tc.want, got)
			}
		})
	}
}
