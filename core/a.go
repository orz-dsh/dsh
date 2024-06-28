package core

import "github.com/orz-dsh/dsh/utils"

type EvalExpr = utils.EvalExpr

type Evaluator = utils.Evaluator

type SystemInfo = utils.SystemInfo

type Logger = utils.Logger

type KVS = utils.DescKeyValues

func kv(key string, value any) utils.DescKeyValue {
	return utils.NewDescKeyValue(key, value)
}

func errN(title string, kvs ...utils.DescKeyValue) error {
	return utils.NewError(1, title, kvs...)
}

func errW(err error, title string, kvs ...utils.DescKeyValue) error {
	return utils.WrapError(1, err, title, kvs...)
}

func desc(title string, kvs ...utils.DescKeyValue) string {
	return utils.NewDesc(title, kvs).String()
}

func reason(reason any) utils.DescKeyValue {
	return kv("reason", reason)
}

func t[T any](expr bool, trueValue T, falseValue T) T {
	return utils.Ternary(expr, trueValue, falseValue)
}

func tfn[T any](expr bool, trueFunc func() T, falseFunc func() T) T {
	return utils.TernaryFunc(expr, trueFunc, falseFunc)
}

func impossible() {
	panic("impossible")
}
