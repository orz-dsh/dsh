package dsh_utils

import "strconv"

func ParseInteger(str string) (int64, error) {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, WrapError(err, "parse integer failed", map[string]any{
			"str": str,
		})
	}
	return value, nil
}

func ParseDecimal(str string) (float64, error) {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, WrapError(err, "parse decimal failed", map[string]any{
			"str": str,
		})
	}
	return value, nil
}
