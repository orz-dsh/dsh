package utils

func TernaryFunc[T any](expr bool, trueFunc func() T, falseFunc func() T) T {
	if expr {
		return trueFunc()
	}
	return falseFunc()
}

func Ternary[T any](expr bool, trueValue T, falseValue T) T {
	if expr {
		return trueValue
	}
	return falseValue
}
