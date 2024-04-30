package dsh_utils

func TernaryFunc[T any](boolExpr bool, trueFunc func() T, falseFunc func() T) T {
	if boolExpr {
		return trueFunc()
	}
	return falseFunc()
}

func Ternary[T any](boolExpr bool, trueValue T, falseValue T) T {
	if boolExpr {
		return trueValue
	}
	return falseValue
}
