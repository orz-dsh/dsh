package dsh_utils

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"strings"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type Error struct {
	Details DescList
	Stacks  errors.StackTrace
	cause   error
}

func (e *Error) Error() string { return e.Details.String() }

func (e *Error) Cause() error { return e.cause }

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		_, _ = io.WriteString(s, "\ndetails:\n")
		_, _ = io.WriteString(s, e.Details.ToString("\t", "\t\t"))
		if e.cause != nil {
			_, _ = io.WriteString(s, "causes:\n")
			causesStr := e.cause.Error()
			causesStr = strings.ReplaceAll(causesStr, "\n", "\n\t")
			_, _ = io.WriteString(s, "\t"+causesStr+"\n")
		}
		if s.Flag('+') {
			_, _ = io.WriteString(s, "stacks:\n")
			for i := 0; i < len(e.Stacks); i++ {
				stackStr := fmt.Sprintf("\t%+v", e.Stacks[i])
				stackStr = strings.ReplaceAll(stackStr, "\n", "\n\t")
				_, _ = io.WriteString(s, stackStr+"\n")
			}
			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, e.Error())
	}
}

func NewError(skip int, title string, kvs ...DescKeyValue) *Error {
	tracer := errors.New("").(stackTracer)
	return &Error{
		Details: DescList{NewDesc(title, kvs)},
		Stacks:  tracer.StackTrace()[skip+1:],
	}
}

func WrapError(skip int, err error, title string, kvs ...DescKeyValue) *Error {
	var err_ *Error
	if errors.As(err, &err_) {
		return &Error{
			Details: append(err_.Details, NewDesc(title, kvs)),
			Stacks:  err_.Stacks,
			cause:   err_.cause,
		}
	}
	if tracer, ok := err.(stackTracer); ok {
		return &Error{
			Details: DescList{NewDesc(title, kvs)},
			Stacks:  tracer.StackTrace(),
			cause:   err,
		}
	}
	tracer := errors.WithStack(err).(stackTracer)
	return &Error{
		Details: DescList{NewDesc(title, kvs)},
		Stacks:  tracer.StackTrace()[skip+1:],
		cause:   err,
	}
}
