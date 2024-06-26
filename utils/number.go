package utils

import "strconv"

func ParseInteger(str string) (int64, error) {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, ErrW(err, "parse integer error", KV("str", str))
	}
	return value, nil
}

func ParseDecimal(str string) (float64, error) {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, ErrW(err, "parse decimal error", KV("str", str))
	}
	return value, nil
}
