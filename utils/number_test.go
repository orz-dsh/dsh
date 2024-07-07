package utils

import "testing"

func TestParseInteger(t *testing.T) {
	integer, err := ParseInt64("123")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(integer)
	integer, err = ParseInt64("")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(integer)
}

func TestParseDecimal(t *testing.T) {
	decimal, err := ParseFloat64("123.1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(decimal)
	decimal, err = ParseFloat64("")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(decimal)
}
