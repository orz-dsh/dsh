package dsh_utils

import (
	"errors"
	errors2 "github.com/pkg/errors"
	"testing"
)

func test11() error {
	return test12()
}

func test12() error {
	return test13()
}

func test13() error {
	return WrapError(
		errors.Join(errors.New("error1"), errors.New("error2"), errors.New("error3")),
		"test error",
		map[string]any{
			"a": "aaa",
		},
	)
}

func test21() error {
	return test22()
}

func test22() error {
	return test23()
}

func test23() error {
	return WrapError(
		errors2.Wrap(errors2.Wrap(errors2.Wrap(errors.New("error1"), "error2"), "error3"), "error4"),
		"test error",
		map[string]any{
			"a": "aaa",
		},
	)
}

func test31() error {
	return test32()
}

func test32() error {
	return test33()
}

func test33() error {
	return NewError(
		"test error",
		map[string]any{
			"a": "aaa",
		},
	)
}

func test41() error {
	return test42()
}

func test42() error {
	return test43()
}

func test43() error {
	return WrapError(
		WrapError(
			NewError(
				"test error1",
				map[string]any{
					"a": "aaa",
					"i": 1,
					"obj": map[string]any{
						"b": "bb\nb",
					},
				},
			),
			"test error2",
			map[string]any{
				"b": "bbb",
			},
		),
		"test error3",
		nil,
	)
}

func TestError1(t *testing.T) {
	err := test11()
	if err != nil {
		t.Logf("%+v", err)
	}
}

func TestError2(t *testing.T) {
	err := test21()
	if err != nil {
		t.Logf("%+v", err)
	}
}

func TestError3(t *testing.T) {
	err := test31()
	if err != nil {
		t.Logf("%+v", err)
	}
}

func TestError4(t *testing.T) {
	err := test41()
	if err != nil {
		t.Logf("%+v", err)
	}
}
