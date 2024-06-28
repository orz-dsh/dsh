package utils

import (
	"crypto/rand"
	"fmt"
)

func RandomString(len int) (string, error) {
	if len%2 != 0 {
		return "", errN("random string error",
			reason("len must be a multiple of 2"),
			kv("len", len),
		)
	}
	bytes := make([]byte, len/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}
