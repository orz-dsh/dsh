package dsh_core

import "dsh/dsh_utils"

func kv(key string, value any) dsh_utils.DescKeyValue {
	return dsh_utils.NewDescKeyValue(key, value)
}

func errN(title string, kvs ...dsh_utils.DescKeyValue) error {
	return dsh_utils.NewError(1, title, kvs...)
}

func errW(err error, title string, kvs ...dsh_utils.DescKeyValue) error {
	return dsh_utils.WrapError(1, err, title, kvs...)
}

func desc(title string, kvs ...dsh_utils.DescKeyValue) string {
	return dsh_utils.NewDesc(title, kvs).String()
}

func reason(reason any) dsh_utils.DescKeyValue {
	return kv("reason", reason)
}

func tfn[T any](expr bool, trueFunc func() T, falseFunc func() T) T {
	return dsh_utils.TernaryFunc(expr, trueFunc, falseFunc)
}
