package dsh_utils

import "testing"

func TestRandomString(t *testing.T) {
	str, err := RandomString(8)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(str)
	}
}
