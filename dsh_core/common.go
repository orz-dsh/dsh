package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
)

type ModelConvertContext struct {
	Title     string
	File      string
	Field     string
	Variables map[string]any
}

func NewModelConvertContext(title, file string) *ModelConvertContext {
	return &ModelConvertContext{
		Title:     title,
		File:      file,
		Variables: map[string]any{},
	}
}

func (c *ModelConvertContext) Child(field string) *ModelConvertContext {
	newField := c.Field
	if newField == "" {
		newField = field
	} else {
		newField += "." + field
	}
	return &ModelConvertContext{
		Title:     c.Title,
		File:      c.File,
		Field:     newField,
		Variables: c.Variables,
	}
}

func (c *ModelConvertContext) ChildItem(field string, index int) *ModelConvertContext {
	return c.Child(fmt.Sprintf("%s[%d]", field, index))
}

func (c *ModelConvertContext) AddVariable(key string, value any) {
	c.Variables[key] = value
}

func (c *ModelConvertContext) GetStringVariable(key string) string {
	if value, exist := c.Variables[key]; exist {
		return value.(string)
	} else {
		impossible()
		return ""
	}
}

func (c *ModelConvertContext) NewError(rsn string, extra ...dsh_utils.DescKeyValue) error {
	kvs := KVS{
		reason(rsn),
		kv("file", c.File),
		kv("field", c.Field),
	}
	return dsh_utils.NewError(1, fmt.Sprintf("%s error", c.Title), append(kvs, extra...)...)
}

func (c *ModelConvertContext) WrapError(err error, rsn string, extra ...dsh_utils.DescKeyValue) error {
	kvs := KVS{
		reason(rsn),
		kv("file", c.File),
		kv("field", c.Field),
	}
	return dsh_utils.WrapError(1, err, fmt.Sprintf("%s error", c.Title), append(kvs, extra...)...)
}

func (c *ModelConvertContext) NewValueEmptyError() error {
	return dsh_utils.NewError(1, fmt.Sprintf("%s error", c.Title),
		reason("value empty"),
		kv("file", c.File),
		kv("field", c.Field),
	)
}

func (c *ModelConvertContext) NewValueInvalidError(value any) error {
	return dsh_utils.NewError(1, fmt.Sprintf("%s error", c.Title),
		reason("value invalid"),
		kv("file", c.File),
		kv("field", c.Field),
		kv("value", value),
	)
}

func (c *ModelConvertContext) WrapValueInvalidError(err error, value any) error {
	return dsh_utils.WrapError(1, err, fmt.Sprintf("%s error", c.Title),
		reason("value invalid"),
		kv("file", c.File),
		kv("field", c.Field),
		kv("value", value),
	)
}
