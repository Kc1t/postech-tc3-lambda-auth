package cpf

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalid = errors.New("cpf invalido")

var nonDigits = regexp.MustCompile(`\D`)

func Normalize(raw string) string {
	return nonDigits.ReplaceAllString(raw, "")
}

func Validate(raw string) (string, error) {
	digits := Normalize(raw)

	if len(digits) != 11 || allSameDigit(digits) {
		return "", ErrInvalid
	}

	if checkDigit(digits, 9) != int(digits[9]-'0') || checkDigit(digits, 10) != int(digits[10]-'0') {
		return "", ErrInvalid
	}

	return digits, nil
}

func allSameDigit(digits string) bool {
	return strings.Count(digits, string(digits[0])) == len(digits)
}

func checkDigit(digits string, position int) int {
	weight := position + 1
	sum := 0

	for i := 0; i < position; i++ {
		sum += int(digits[i]-'0') * weight
		weight--
	}

	remainder := sum % 11
	if remainder < 2 {
		return 0
	}

	return 11 - remainder
}
