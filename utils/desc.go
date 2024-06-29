package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// region Desc

type Desc struct {
	Title string
	Body  DescBody
}

type DescList []Desc

func NewDesc(title string, kvs DescKeyValues) Desc {
	return Desc{
		Title: title,
		Body:  NewDescBody(kvs),
	}
}

func (d Desc) ToString(titleIdent string, bodyIdent string) string {
	if len(d.Body) == 0 {
		return fmt.Sprintf("%s%s\n", titleIdent, d.Title)
	}
	return fmt.Sprintf("%s%s\n%s", titleIdent, d.Title, d.Body.ToString(bodyIdent))
}

func (d Desc) String() string {
	return d.ToString("", "\t")
}

func (l DescList) ToString(titleIdent string, bodyIdent string) string {
	var builder strings.Builder
	for i := 0; i < len(l); i++ {
		builder.WriteString(l[i].ToString(titleIdent, bodyIdent))
	}
	return builder.String()
}

func (l DescList) String() string {
	return l.ToString("", "\t")
}

// endregion

// region DescBody

type DescBody []string

func NewDescBody(kvs DescKeyValues) DescBody {
	return newDescBody("", kvs, map[uintptr]string{}, true, 0, 20)
}

func newDescBody(keyPrefix string, kvs DescKeyValues, pointers map[uintptr]string, nextLevel bool, depth int, maxDepth int) DescBody {
	var body DescBody
	if depth >= maxDepth {
		for i := 0; i < len(kvs); i++ {
			body = append(body, kvs[i].ToString(keyPrefix, "<omit>"))
		}
		return body
	}
	for i := 0; i < len(kvs); i++ {
		item := kvs[i]
		link := ""
		var itemBody DescBody
		if item.valueIsNil {
			itemBody = DescBody{item.ToString(keyPrefix, "<nil>")}
		} else {
			if descKeyValuesFunc, ok := item.Value.(DescKeyValuesFunc); ok {
				switch item.valueReflect.Kind() {
				case reflect.Slice, reflect.Map, reflect.Pointer:
					if key, exist := pointers[item.valueReflect.Pointer()]; exist {
						link = key
					} else {
						pointers[item.valueReflect.Pointer()] = keyPrefix + item.Key
					}
				default:
				}
				if link == "" {
					newKvs := descKeyValuesFunc.DescKeyValues()
					if len(newKvs) > 0 {
						itemBody = newDescBody(keyPrefix+item.Key+".", newKvs, pointers, true, depth+1, maxDepth)
					}
				}
			} else {
				switch item.valueReflect.Kind() {
				case reflect.Slice:
					if key, exist := pointers[item.valueReflect.Pointer()]; exist {
						link = key
					} else {
						pointers[item.valueReflect.Pointer()] = keyPrefix + item.Key
					}
					fallthrough
				case reflect.Array:
					if link == "" && item.valueReflect.Len() > 0 {
						newKvs := make(DescKeyValues, item.valueReflect.Len())
						for j := 0; j < item.valueReflect.Len(); j++ {
							newKvs[j] = NewDescKeyValue(fmt.Sprintf("[%d]", j), item.valueReflect.Index(j).Interface())
						}
						itemBody = newDescBody(keyPrefix+item.Key, newKvs, pointers, true, depth+1, maxDepth)
					}
				case reflect.Map:
					if key, exist := pointers[item.valueReflect.Pointer()]; exist {
						link = key
					} else {
						pointers[item.valueReflect.Pointer()] = keyPrefix + item.Key
						keys := item.valueReflect.MapKeys()
						if len(keys) > 0 {
							newKvs := make(DescKeyValues, len(keys))
							for j := 0; j < len(keys); j++ {
								newKvs[j] = NewDescKeyValue(fmt.Sprintf("[`%s`]", keys[j]), item.valueReflect.MapIndex(keys[j]).Interface())
							}
							itemBody = newDescBody(keyPrefix+item.Key, newKvs, pointers, true, depth+1, maxDepth)
						}
					}
				case reflect.Pointer:
					if key, exist := pointers[item.valueReflect.Pointer()]; exist {
						link = key
					} else {
						pointers[item.valueReflect.Pointer()] = keyPrefix + item.Key
						newKvs := DescKeyValues{NewDescKeyValue(item.Key, item.valueReflect.Elem().Interface())}
						itemBody = newDescBody(keyPrefix, newKvs, pointers, false, depth, maxDepth)
					}
				case reflect.Struct:
					valueType := item.valueReflect.Type()
					var newKvs DescKeyValues
					for j := 0; j < valueType.NumField(); j++ {
						field := valueType.Field(j)
						if field.IsExported() && field.Tag.Get("desc") != "-" {
							newKvs = append(newKvs, NewDescKeyValue(field.Name, item.valueReflect.Field(j).Interface()))
						}
					}
					if len(newKvs) > 0 {
						itemBody = newDescBody(keyPrefix+item.Key+".", newKvs, pointers, true, depth+1, maxDepth)
					}
				default:
					itemBody = DescBody{item.ToString(keyPrefix, "")}
				}
			}
			if link == "" {
				if descExtraKeyValuesFunc, ok := item.Value.(DescExtraKeyValuesFunc); ok {
					extraBody := newDescBody(keyPrefix+item.Key+".", descExtraKeyValuesFunc.DescExtraKeyValues(), pointers, true, depth+1, maxDepth)
					itemBody = append(itemBody, extraBody...)
				}
			}
		}
		if link != "" {
			body = append(body, item.ToString(keyPrefix, fmt.Sprintf("<link:%s>", link)))
		} else if len(itemBody) > 0 {
			body = append(body, itemBody...)
		} else if nextLevel {
			body = append(body, item.ToString(keyPrefix, "<empty>"))
		}
	}
	return body
}

func (b DescBody) ToString(ident string) string {
	var builder strings.Builder
	for i := 0; i < len(b); i++ {
		builder.WriteString(ident + b[i] + "\n")
	}
	return builder.String()
}

func (b DescBody) String() string {
	return b.ToString("")
}

// endregion

// region DescKeyValue

type DescKeyValue struct {
	Key          string
	Value        any
	valueReflect reflect.Value
	valueIsNil   bool
}

type DescKeyValues []DescKeyValue

type DescKeyValuesFunc interface {
	DescKeyValues() DescKeyValues
}

type DescExtraKeyValuesFunc interface {
	DescExtraKeyValues() DescKeyValues
}

func NewDescKeyValue(key string, value any) DescKeyValue {
	switch value.(type) {
	case reflect.Type:
		return NewDescKeyValue(key, value.(reflect.Type).String())
	case reflect.Value:
		return NewDescKeyValue(key, value.(reflect.Value).String())
	case reflect.Kind:
		return NewDescKeyValue(key, value.(reflect.Kind).String())
	}
	valueReflect := reflect.ValueOf(value)
	valueIsNil := value == nil
	switch valueReflect.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		valueIsNil = valueReflect.IsNil()
	default:
	}
	return DescKeyValue{
		Key:          key,
		Value:        value,
		valueReflect: valueReflect,
		valueIsNil:   valueIsNil,
	}
}

func (kv DescKeyValue) ToString(keyPrefix string, valueLabel string) string {
	if valueLabel != "" {
		return fmt.Sprintf("%s%s = %s", keyPrefix, kv.Key, valueLabel)
	} else {
		switch kv.valueReflect.Kind() {
		case reflect.Func:
			var inTypes strings.Builder
			var outTypes strings.Builder
			valueType := kv.valueReflect.Type()
			for j := 0; j < valueType.NumIn(); j++ {
				if j > 0 {
					inTypes.WriteString(", ")
				}
				inTypes.WriteString(valueType.In(j).String())
			}
			for j := 0; j < valueType.NumOut(); j++ {
				if j > 0 {
					outTypes.WriteString(", ")
				}
				outTypes.WriteString(valueType.Out(j).String())
			}
			return fmt.Sprintf("%s%s = `%v - func(%s) (%s)`", keyPrefix, kv.Key, kv.Value, inTypes.String(), outTypes.String())
		case reflect.Chan:
			valueType := kv.valueReflect.Type()
			return fmt.Sprintf("%s%s = `%v - %s %s`", keyPrefix, kv.Key, kv.Value, valueType.ChanDir(), valueType.Elem().String())
		case reflect.UnsafePointer:
			return fmt.Sprintf("%s%s = `%v - unsafe.Pointer`", keyPrefix, kv.Key, kv.Value)
		case reflect.Uintptr:
			return fmt.Sprintf("%s%s = `%#x - uintptr`", keyPrefix, kv.Key, kv.Value)
		default:
			vStr := fmt.Sprintf("%+v", kv.Value)
			vStr = strings.ReplaceAll(vStr, "\n", "\\n")
			vStr = strings.ReplaceAll(vStr, "\r", "\\r")
			return fmt.Sprintf("%s%s = `%s`", keyPrefix, kv.Key, vStr)
		}
	}
}

func (kv DescKeyValue) String() string {
	return kv.ToString("", "")
}

func (kv DescKeyValue) DescKeyValues() DescKeyValues {
	return DescKeyValues{kv}
}

func (kvs DescKeyValues) DescKeyValues() DescKeyValues {
	return kvs
}

// endregion
