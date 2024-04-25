package dsh_utils

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"strings"
)

type Message struct {
	Title string
	Body  MessageBody
}

type MessageBody []string

type Messages []Message

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func (message Message) ToString(ident string) string {
	if len(message.Body) == 0 {
		return fmt.Sprintf("%s%s\n", ident, message.Title)
	}
	return fmt.Sprintf("%s%s\n%s", ident, message.Title, message.Body.ToString(ident+"\t"))
}

func (message Message) String() string {
	return message.ToString("")
}

func (body MessageBody) ToString(ident string) string {
	var builder strings.Builder
	for i := 0; i < len(body); i++ {
		builder.WriteString(ident + body[i] + "\n")
	}
	return builder.String()
}

func (body MessageBody) String() string {
	return body.ToString("")
}

func (messages Messages) ToString(ident string) string {
	var builder strings.Builder
	for i := 0; i < len(messages); i++ {
		builder.WriteString(messages[i].ToString(ident))
	}
	return builder.String()
}

func (messages Messages) String() string {
	return messages.ToString("")
}

type Error struct {
	Messages Messages
	Stacks   errors.StackTrace
	cause    error
}

func (e *Error) Error() string { return e.Messages.String() }

func (e *Error) Cause() error { return e.cause }

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		_, _ = io.WriteString(s, "\nmessages:\n")
		_, _ = io.WriteString(s, e.Messages.ToString("\t"))
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

func NewMessageBody(bodyMap map[string]interface{}) MessageBody {
	var body []string
	for k, v := range bodyMap {
		vStr := fmt.Sprintf("%v", v)
		vStr = strings.ReplaceAll(vStr, "\n", "\\n")
		vStr = strings.ReplaceAll(vStr, "\r", "\\r")
		body = append(body, k+" = "+vStr)
	}
	return body
}

func NewError(title string, body map[string]interface{}) error {
	tracer := errors.New("").(stackTracer)
	return &Error{
		Messages: Messages{
			{Title: title, Body: NewMessageBody(body)},
		},
		Stacks: tracer.StackTrace()[1:],
	}
}

func WrapError(err error, title string, body map[string]interface{}) error {
	var err_ *Error
	if errors.As(err, &err_) {
		return &Error{
			Messages: append(err_.Messages, Message{Title: title, Body: NewMessageBody(body)}),
			Stacks:   err_.Stacks,
			cause:    err_.cause,
		}
	}
	if tracer, ok := err.(stackTracer); ok {
		return &Error{
			Messages: Messages{
				{Title: title, Body: NewMessageBody(body)},
			},
			Stacks: tracer.StackTrace(),
			cause:  err,
		}
	}
	tracer := errors.WithStack(err).(stackTracer)
	return &Error{
		Messages: Messages{
			{Title: title, Body: NewMessageBody(body)},
		},
		Stacks: tracer.StackTrace()[1:],
		cause:  err,
	}
}
