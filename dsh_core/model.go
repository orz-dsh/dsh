package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
)

type modelHelper struct {
	Title     string
	File      string
	Field     string
	Variables map[string]any
}

func newModelHelper(title, file string) *modelHelper {
	return &modelHelper{
		Title:     title,
		File:      file,
		Variables: map[string]any{},
	}
}

func (h *modelHelper) Child(field string) *modelHelper {
	newField := h.Field
	if newField == "" {
		newField = field
	} else {
		newField += "." + field
	}
	return &modelHelper{
		Title:     h.Title,
		File:      h.File,
		Field:     newField,
		Variables: h.Variables,
	}
}

func (h *modelHelper) ChildItem(field string, index int) *modelHelper {
	return h.Child(fmt.Sprintf("%s[%d]", field, index))
}

func (h *modelHelper) Item(index int) *modelHelper {
	return &modelHelper{
		Title:     h.Title,
		File:      h.File,
		Field:     fmt.Sprintf("%s[%d]", h.Field, index),
		Variables: h.Variables,
	}
}

func (h *modelHelper) AddVariable(key string, value any) *modelHelper {
	h.Variables[key] = value
	return h
}

func (h *modelHelper) GetStringVariable(key string) string {
	if value, exist := h.Variables[key]; exist {
		return value.(string)
	} else {
		impossible()
		return ""
	}
}

func (h *modelHelper) NewError(rsn string, extra ...dsh_utils.DescKeyValue) error {
	kvs := KVS{
		reason(rsn),
		kv("file", h.File),
		kv("field", h.Field),
	}
	return dsh_utils.NewError(1, fmt.Sprintf("%s error", h.Title), append(kvs, extra...)...)
}

func (h *modelHelper) WrapError(err error, rsn string, extra ...dsh_utils.DescKeyValue) error {
	kvs := KVS{
		reason(rsn),
		kv("file", h.File),
		kv("field", h.Field),
	}
	return dsh_utils.WrapError(1, err, fmt.Sprintf("%s error", h.Title), append(kvs, extra...)...)
}

func (h *modelHelper) NewValueEmptyError() error {
	return dsh_utils.NewError(1, fmt.Sprintf("%s error", h.Title),
		reason("value empty"),
		kv("file", h.File),
		kv("field", h.Field),
	)
}

func (h *modelHelper) NewValueInvalidError(value any) error {
	return dsh_utils.NewError(1, fmt.Sprintf("%s error", h.Title),
		reason("value invalid"),
		kv("file", h.File),
		kv("field", h.Field),
		kv("value", value),
	)
}

func (h *modelHelper) WrapValueInvalidError(err error, value any) error {
	return dsh_utils.WrapError(1, err, fmt.Sprintf("%s error", h.Title),
		reason("value invalid"),
		kv("file", h.File),
		kv("field", h.Field),
		kv("value", value),
	)
}

func (h *modelHelper) ConvertEvalExpr(field, expr string) (*EvalExpr, error) {
	if expr != "" {
		exprObj, err := dsh_utils.CompileExpr(expr)
		if err != nil {
			return nil, h.Child(field).WrapValueInvalidError(err, expr)
		}
		return exprObj, nil
	}
	return nil, nil
}

func (h *modelHelper) CheckStringItemEmpty(field string, items []string) error {
	for i := 0; i < len(items); i++ {
		if items[i] == "" {
			return h.ChildItem(field, i).NewValueEmptyError()
		}
	}
	return nil
}

type model[R any] interface {
	convert(helper *modelHelper) (R, error)
}

func convertChildModels[R any, M model[R]](helper *modelHelper, field string, models []M) ([]R, error) {
	var result []R
	for i := 0; i < len(models); i++ {
		r, err := models[i].convert(helper.ChildItem(field, i))
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
