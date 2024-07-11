package utils

import (
	"fmt"
)

// region Model

type Model[R any] interface {
	Convert(helper *ModelHelper) (R, error)
}

// endregion

// region ModelHelper

type ModelHelper struct {
	Logger    *Logger
	Title     string
	Source    string
	Field     string
	Variables map[string]any
}

func NewModelHelper(logger *Logger, title, source string) *ModelHelper {
	return &ModelHelper{
		Logger:    logger,
		Title:     title,
		Source:    source,
		Variables: map[string]any{},
	}
}

func (h *ModelHelper) Child(field string) *ModelHelper {
	newField := h.Field
	if newField == "" {
		newField = field
	} else {
		newField += "." + field
	}
	return &ModelHelper{
		Logger:    h.Logger,
		Title:     h.Title,
		Source:    h.Source,
		Field:     newField,
		Variables: h.Variables,
	}
}

func (h *ModelHelper) ChildItem(field string, index int) *ModelHelper {
	return h.Child(fmt.Sprintf("%s[%d]", field, index))
}

func (h *ModelHelper) Item(index int) *ModelHelper {
	return &ModelHelper{
		Logger:    h.Logger,
		Title:     h.Title,
		Source:    h.Source,
		Field:     fmt.Sprintf("%s[%d]", h.Field, index),
		Variables: h.Variables,
	}
}

func (h *ModelHelper) AddVariable(key string, value any) *ModelHelper {
	h.Variables[key] = value
	return h
}

func (h *ModelHelper) GetStringVariable(key string) string {
	if value, exist := h.Variables[key]; exist {
		return value.(string)
	} else {
		Impossible()
		return ""
	}
}

func (h *ModelHelper) WarnValueUnsound(value any) {
	h.Logger.WarnDesc(fmt.Sprintf("%s warn", h.Title),
		Reason("value unsound"),
		KV("source", h.Source),
		KV("field", h.Field),
		KV("value", value),
	)
}

func (h *ModelHelper) WarnValueUseless(value any) {
	h.Logger.WarnDesc(fmt.Sprintf("%s warn", h.Title),
		Reason("value useless"),
		KV("source", h.Source),
		KV("field", h.Field),
		KV("value", value),
	)
}

func (h *ModelHelper) Warn(reason string, extra ...DescKeyValue) {
	kvs := KVS{
		Reason(reason),
		KV("source", h.Source),
		KV("field", h.Field),
	}
	h.Logger.WarnDesc(fmt.Sprintf("%s warn", h.Title), append(kvs, extra...)...)
}

func (h *ModelHelper) NewError(reason string, extra ...DescKeyValue) error {
	kvs := KVS{
		Reason(reason),
		KV("source", h.Source),
		KV("field", h.Field),
	}
	return NewError(1, fmt.Sprintf("%s error", h.Title), append(kvs, extra...)...)
}

func (h *ModelHelper) WrapError(err error, reason string, extra ...DescKeyValue) error {
	kvs := KVS{
		Reason(reason),
		KV("source", h.Source),
		KV("field", h.Field),
	}
	return WrapError(1, err, fmt.Sprintf("%s error", h.Title), append(kvs, extra...)...)
}

func (h *ModelHelper) NewValueEmptyError() error {
	return NewError(1, fmt.Sprintf("%s error", h.Title),
		Reason("value empty"),
		KV("source", h.Source),
		KV("field", h.Field),
	)
}

func (h *ModelHelper) NewValueInvalidError(value any) error {
	return NewError(1, fmt.Sprintf("%s error", h.Title),
		Reason("value invalid"),
		KV("source", h.Source),
		KV("field", h.Field),
		KV("value", value),
	)
}

func (h *ModelHelper) WrapValueInvalidError(err error, value any) error {
	return WrapError(1, err, fmt.Sprintf("%s error", h.Title),
		Reason("value invalid"),
		KV("source", h.Source),
		KV("field", h.Field),
		KV("value", value),
	)
}

func (h *ModelHelper) CheckStringItemEmpty(field string, items []string) error {
	for i := 0; i < len(items); i++ {
		if items[i] == "" {
			return h.ChildItem(field, i).NewValueEmptyError()
		}
	}
	return nil
}

func ConvertChildModels[R any, M Model[R]](helper *ModelHelper, field string, models []M) ([]R, error) {
	var result []R
	for i := 0; i < len(models); i++ {
		r, err := models[i].Convert(helper.ChildItem(field, i))
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// endregion
