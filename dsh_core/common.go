package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
)

type modelConvertContext struct {
	Title     string
	File      string
	Field     string
	Variables map[string]any
}

func newModelConvertContext(title, file string) *modelConvertContext {
	return &modelConvertContext{
		Title:     title,
		File:      file,
		Variables: map[string]any{},
	}
}

func (c *modelConvertContext) Child(field string) *modelConvertContext {
	newField := c.Field
	if newField == "" {
		newField = field
	} else {
		newField += "." + field
	}
	return &modelConvertContext{
		Title:     c.Title,
		File:      c.File,
		Field:     newField,
		Variables: c.Variables,
	}
}

func (c *modelConvertContext) ChildItem(field string, index int) *modelConvertContext {
	return c.Child(fmt.Sprintf("%s[%d]", field, index))
}

func (c *modelConvertContext) AddVariable(key string, value any) {
	c.Variables[key] = value
}

func (c *modelConvertContext) GetStringVariable(key string) string {
	if value, exist := c.Variables[key]; exist {
		return value.(string)
	} else {
		impossible()
		return ""
	}
}

func (c *modelConvertContext) NewError(rsn string, extra ...dsh_utils.DescKeyValue) error {
	kvs := KVS{
		reason(rsn),
		kv("file", c.File),
		kv("field", c.Field),
	}
	return dsh_utils.NewError(1, fmt.Sprintf("%s error", c.Title), append(kvs, extra...)...)
}

func (c *modelConvertContext) WrapError(err error, rsn string, extra ...dsh_utils.DescKeyValue) error {
	kvs := KVS{
		reason(rsn),
		kv("file", c.File),
		kv("field", c.Field),
	}
	return dsh_utils.WrapError(1, err, fmt.Sprintf("%s error", c.Title), append(kvs, extra...)...)
}

func (c *modelConvertContext) NewValueEmptyError() error {
	return dsh_utils.NewError(1, fmt.Sprintf("%s error", c.Title),
		reason("value empty"),
		kv("file", c.File),
		kv("field", c.Field),
	)
}

func (c *modelConvertContext) NewValueInvalidError(value any) error {
	return dsh_utils.NewError(1, fmt.Sprintf("%s error", c.Title),
		reason("value invalid"),
		kv("file", c.File),
		kv("field", c.Field),
		kv("value", value),
	)
}

func (c *modelConvertContext) WrapValueInvalidError(err error, value any) error {
	return dsh_utils.WrapError(1, err, fmt.Sprintf("%s error", c.Title),
		reason("value invalid"),
		kv("file", c.File),
		kv("field", c.Field),
		kv("value", value),
	)
}
