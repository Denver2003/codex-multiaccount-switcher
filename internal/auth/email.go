package auth

import (
	"regexp"
)

var basicEmailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func ExtractEmail(raw []byte) (string, error) {
	value, err := parseJSONObject(raw)
	if err != nil {
		return "", err
	}

	emailValue, ok := value["email"]
	if !ok {
		return "", nil
	}

	email, ok := emailValue.(string)
	if !ok {
		return "", nil
	}

	if !basicEmailPattern.MatchString(email) {
		return "", nil
	}

	return email, nil
}
