package utils

import (
	"maps"
	"reflect"
)

type MapMergeMode string

const (
	MapMergeModeReplace MapMergeMode = "replace"
	MapMergeModeInsert  MapMergeMode = "insert"
	MapMergeModeRootKey string       = "$root"
)

func MapAnyByStr[E any, M map[string]E](m M) map[string]any {
	result := map[string]any{}
	for k, v := range m {
		result[k] = v
	}
	return result
}

func MapCopy[K comparable, V any](source map[K]V) map[K]V {
	if source == nil {
		return map[K]V{}
	}
	result := map[K]V{}
	maps.Copy(result, source)
	return result
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	if m == nil {
		return nil
	}
	result := make([]K, 0, len(m))
	for k, _ := range m {
		result = append(result, k)
	}
	return result
}

func MapValues[K comparable, V any](m map[K]V) []V {
	if m == nil {
		return nil
	}
	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func MapMerge(target map[string]any, source map[string]any, merge map[string]MapMergeMode, label string, trace map[string]any) (map[string]any, map[string]any, error) {
	if merge == nil {
		merge = map[string]MapMergeMode{}
	}
	if target != nil && merge[MapMergeModeRootKey] == MapMergeModeReplace {
		clear(target)
	}
	tracer := newMapMergeTracer(label, trace)
	target, err := mapMerge(target, source, merge, tracer, "")
	if err != nil {
		return nil, nil, err
	}
	return target, tracer.Trace, nil
}

func mapMerge(target map[string]any, source map[string]any, merge map[string]MapMergeMode, tracer *mapMergeTracer, parent string) (map[string]any, error) {
	if target == nil {
		target = map[string]any{}
	}
	for key, sourceValue := range source {
		field := key
		if parent != "" {
			field = parent + "." + key
		}
		targetValue := target[key]
		switch sourceValue.(type) {
		case map[string]any:
			sourceMap := sourceValue.(map[string]any)
			if targetValue == nil {
				if targetResult, err := mapMerge(nil, sourceMap, merge, tracer.empty(key), field); err != nil {
					return nil, err
				} else {
					target[key] = targetResult
				}
			} else if targetMap, ok := targetValue.(map[string]any); ok {
				if mode, exist := merge[field]; exist {
					switch mode {
					case MapMergeModeReplace:
						if targetResult, err := mapMerge(nil, sourceMap, merge, tracer.empty(key), field); err != nil {
							return nil, err
						} else {
							target[key] = targetResult
						}
					default:
						return nil, ErrN("merge map error",
							Reason("merge mode invalid"),
							KV("field", field),
							KV("specifyMode", mode),
							KV("supportModes", []MapMergeMode{
								MapMergeModeReplace,
							}),
						)
					}
				} else {
					if _, err := mapMerge(targetMap, sourceMap, merge, tracer.child(key), field); err != nil {
						return nil, err
					}
				}
			} else {
				return nil, ErrN("merge map error",
					Reason("value type not match"),
					KV("field", field),
					KV("sourceType", reflect.TypeOf(sourceValue)),
					KV("targetType", reflect.TypeOf(targetValue)),
				)
			}
		case []any:
			sourceList := sourceValue.([]any)
			if targetValue == nil {
				target[key] = sourceList
				tracer.addNewList(key, len(sourceList))
			} else if targetList, ok := targetValue.([]any); ok {
				if mode, exist := merge[field]; exist {
					if mode == MapMergeModeReplace {
						target[key] = sourceList
						tracer.addNewList(key, len(sourceList))
					} else if mode == MapMergeModeInsert {
						target[key] = append(sourceList, targetList...)
						tracer.addInsertList(key, len(sourceList))
					} else {
						return nil, ErrN("merge map error",
							Reason("merge type invalid"),
							KV("field", field),
							KV("specifyMode", mode),
							KV("supportModes", []MapMergeMode{
								MapMergeModeReplace,
								MapMergeModeInsert,
							}),
						)
					}
				} else {
					target[key] = append(targetList, sourceList...)
					tracer.addAppendList(key, len(sourceList))
				}
			} else {
				return nil, ErrN("merge map error",
					Reason("value type not match"),
					KV("field", field),
					KV("sourceType", reflect.TypeOf(sourceValue)),
					KV("targetType", reflect.TypeOf(targetValue)),
				)
			}
		default:
			if sourceValue != nil {
				switch targetValue.(type) {
				case map[string]any:
					return nil, ErrN("merge map error",
						Reason("value type not match"),
						KV("field", field),
						KV("sourceType", reflect.TypeOf(sourceValue)),
						KV("targetType", reflect.TypeOf(targetValue)),
					)
				case []any:
					return nil, ErrN("merge map error",
						Reason("value type not match"),
						KV("field", field),
						KV("sourceType", reflect.TypeOf(sourceValue)),
						KV("targetType", reflect.TypeOf(targetValue)),
					)
				}
			}
			target[key] = sourceValue
			tracer.add(key)
		}
	}
	return target, nil
}

// region mapMergeTracer

type mapMergeTracer struct {
	Label string
	Trace map[string]any
}

func newMapMergeTracer(label string, trace map[string]any) *mapMergeTracer {
	if trace == nil {
		trace = map[string]any{}
	}
	return &mapMergeTracer{
		Label: label,
		Trace: trace,
	}
}

func (t *mapMergeTracer) empty(key string) *mapMergeTracer {
	childTrace := map[string]any{}
	t.Trace[key] = childTrace
	return newMapMergeTracer(t.Label, childTrace)
}

func (t *mapMergeTracer) child(key string) *mapMergeTracer {
	if childTrace, ok := t.Trace[key].(map[string]any); ok {
		return newMapMergeTracer(t.Label, childTrace)
	}
	childTrace := map[string]any{}
	t.Trace[key] = childTrace
	return newMapMergeTracer(t.Label, childTrace)
}

func (t *mapMergeTracer) add(key string) {
	t.Trace[key] = t.Label
}

func (t *mapMergeTracer) addNewList(key string, len int) {
	var list []any
	for i := 0; i < len; i++ {
		list = append(list, t.Label)
	}
	t.Trace[key] = list
}

func (t *mapMergeTracer) addInsertList(key string, len int) {
	var list []any
	for i := 0; i < len; i++ {
		list = append(list, t.Label)
	}
	if existList, ok := t.Trace[key].([]any); ok {
		t.Trace[key] = append(list, existList...)
	} else {
		t.Trace[key] = list
	}
}

func (t *mapMergeTracer) addAppendList(key string, len int) {
	var list []any
	for i := 0; i < len; i++ {
		list = append(list, t.Label)
	}
	if existList, ok := t.Trace[key].([]any); ok {
		t.Trace[key] = append(existList, list...)
	} else {
		t.Trace[key] = list
	}
}

// endregion
