package dsh_utils

func kv(key string, value any) DescKeyValue {
	return NewDescKeyValue(key, value)
}

func errN(title string, kvs ...DescKeyValue) error {
	return NewError(1, title, kvs...)
}

func errW(err error, title string, kvs ...DescKeyValue) error {
	return WrapError(1, err, title, kvs...)
}

func desc(title string, kvs ...DescKeyValue) string {
	return NewDesc(title, kvs).String()
}

func reason(reason any) DescKeyValue {
	return kv("reason", reason)
}

func tfn[T any](expr bool, trueFunc func() T, falseFunc func() T) T {
	return TernaryFunc(expr, trueFunc, falseFunc)
}