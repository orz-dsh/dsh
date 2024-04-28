package dsh_utils

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"strings"
)

type ErrorDetail struct {
	Title string
	Body  ErrorDetailBody
}

type ErrorDetailBody []string

type ErrorDetails []ErrorDetail

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func (detail ErrorDetail) ToString(ident string) string {
	if len(detail.Body) == 0 {
		return fmt.Sprintf("%s%s\n", ident, detail.Title)
	}
	return fmt.Sprintf("%s%s\n%s", ident, detail.Title, detail.Body.ToString(ident+"\t"))
}

func (detail ErrorDetail) String() string {
	return detail.ToString("")
}

func NewErrorDetailBody(bodyMap map[string]any) ErrorDetailBody {
	var body []string
	for k, v := range bodyMap {
		vStr := fmt.Sprintf("%v", v)
		vStr = strings.ReplaceAll(vStr, "\n", "\\n")
		vStr = strings.ReplaceAll(vStr, "\r", "\\r")
		body = append(body, k+" = `"+vStr+"`")
	}
	return body
}

func (body ErrorDetailBody) ToString(ident string) string {
	var builder strings.Builder
	for i := 0; i < len(body); i++ {
		builder.WriteString(ident + body[i] + "\n")
	}
	return builder.String()
}

func (body ErrorDetailBody) String() string {
	return body.ToString("")
}

func (details ErrorDetails) ToString(ident string) string {
	var builder strings.Builder
	for i := 0; i < len(details); i++ {
		builder.WriteString(details[i].ToString(ident))
	}
	return builder.String()
}

func (details ErrorDetails) String() string {
	return details.ToString("")
}

type Error struct {
	Details ErrorDetails
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
		_, _ = io.WriteString(s, e.Details.ToString("\t"))
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

func NewError(title string, body map[string]any) error {
	tracer := errors.New("").(stackTracer)
	return &Error{
		Details: ErrorDetails{
			{Title: title, Body: NewErrorDetailBody(body)},
		},
		Stacks: tracer.StackTrace()[1:],
	}
}

func WrapError(err error, title string, body map[string]any) error {
	var err_ *Error
	if errors.As(err, &err_) {
		return &Error{
			Details: append(err_.Details, ErrorDetail{Title: title, Body: NewErrorDetailBody(body)}),
			Stacks:  err_.Stacks,
			cause:   err_.cause,
		}
	}
	if tracer, ok := err.(stackTracer); ok {
		return &Error{
			Details: ErrorDetails{
				{Title: title, Body: NewErrorDetailBody(body)},
			},
			Stacks: tracer.StackTrace(),
			cause:  err,
		}
	}
	tracer := errors.WithStack(err).(stackTracer)
	return &Error{
		Details: ErrorDetails{
			{Title: title, Body: NewErrorDetailBody(body)},
		},
		Stacks: tracer.StackTrace()[1:],
		cause:  err,
	}
}
