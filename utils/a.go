package utils

type KVS = DescKeyValues

func KV(key string, value any) DescKeyValue {
	return NewDescKeyValue(key, value)
}

func ErrN(title string, kvs ...DescKeyValue) error {
	return NewError(1, title, kvs...)
}

func ErrW(err error, title string, kvs ...DescKeyValue) error {
	return WrapError(1, err, title, kvs...)
}

func DescN(title string, kvs ...DescKeyValue) string {
	return NewDesc(title, kvs).String()
}

func Reason(reason any) DescKeyValue {
	return KV("reason", reason)
}

func ValT[T any](expr bool, trueValue T, falseValue T) T {
	return Ternary(expr, trueValue, falseValue)
}

func FuncT[T any](expr bool, trueFunc func() T, falseFunc func() T) T {
	return TernaryFunc(expr, trueFunc, falseFunc)
}

func Impossible() {
	panic("impossible")
}
