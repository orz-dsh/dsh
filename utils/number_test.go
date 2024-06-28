package utils

import "testing"

func TestParseInteger(t *testing.T) {
	integer, err := ParseInteger("123")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(integer)
	integer, err = ParseInteger("")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(integer)
}

func TestParseDecimal(t *testing.T) {
	decimal, err := ParseDecimal("123.1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(decimal)
	decimal, err = ParseDecimal("")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(decimal)
}
