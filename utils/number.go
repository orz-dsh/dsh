package utils

import "strconv"

func ParseInt32(str string) (int, error) {
	value, err := strconv.Atoi(str)
	if err != nil {
		return 0, ErrW(err, "parse int32 error", KV("str", str))
	}
	return value, nil
}

func ParseInt64(str string) (int64, error) {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, ErrW(err, "parse int64 error", KV("str", str))
	}
	return value, nil
}

func ParseFloat32(str string) (float32, error) {
	value, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return 0, ErrW(err, "parse decimal error", KV("str", str))
	}
	return float32(value), nil
}

func ParseFloat64(str string) (float64, error) {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, ErrW(err, "parse decimal error", KV("str", str))
	}
	return value, nil
}
