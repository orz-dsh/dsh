package utils

import "reflect"

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

func MapMerge(target map[string]any, source map[string]any, modes map[string]MapMergeMode, label string, traces map[string]any) (map[string]any, map[string]any, error) {
	if modes == nil {
		modes = map[string]MapMergeMode{}
	}
	if target != nil && modes[MapMergeModeRootKey] == MapMergeModeReplace {
		clear(target)
	}
	tracer := newMapMergeTracer(label, traces)
	target, err := mapMerge(target, source, modes, tracer, "")
	if err != nil {
		return nil, nil, err
	}
	return target, tracer.Traces, nil
}

func mapMerge(target map[string]any, source map[string]any, modes map[string]MapMergeMode, tracer *mapMergeTracer, parent string) (map[string]any, error) {
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
				if targetResult, err := mapMerge(nil, sourceMap, modes, tracer.empty(key), field); err != nil {
					return nil, err
				} else {
					target[key] = targetResult
				}
			} else if targetMap, ok := targetValue.(map[string]any); ok {
				if mode, exist := modes[field]; exist {
					switch mode {
					case MapMergeModeReplace:
						if targetResult, err := mapMerge(nil, sourceMap, modes, tracer.empty(key), field); err != nil {
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
					if _, err := mapMerge(targetMap, sourceMap, modes, tracer.child(key), field); err != nil {
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
				tracer.traceNewList(key, len(sourceList))
			} else if targetList, ok := targetValue.([]any); ok {
				if mode, exist := modes[field]; exist {
					if mode == MapMergeModeReplace {
						target[key] = sourceList
						tracer.traceNewList(key, len(sourceList))
					} else if mode == MapMergeModeInsert {
						target[key] = append(sourceList, targetList...)
						tracer.traceInsertList(key, len(sourceList))
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
					tracer.traceAppendList(key, len(sourceList))
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
			tracer.trace(key)
		}
	}
	return target, nil
}

// region mapMergeTracer

type mapMergeTracer struct {
	Label  string
	Traces map[string]any
}

func newMapMergeTracer(label string, traces map[string]any) *mapMergeTracer {
	if traces == nil {
		traces = map[string]any{}
	}
	return &mapMergeTracer{
		Label:  label,
		Traces: traces,
	}
}

func (t *mapMergeTracer) empty(key string) *mapMergeTracer {
	childTraces := map[string]any{}
	t.Traces[key] = childTraces
	return newMapMergeTracer(t.Label, childTraces)
}

func (t *mapMergeTracer) child(key string) *mapMergeTracer {
	if childTraces, ok := t.Traces[key].(map[string]any); ok {
		return newMapMergeTracer(t.Label, childTraces)
	}
	childTrace := map[string]any{}
	t.Traces[key] = childTrace
	return newMapMergeTracer(t.Label, childTrace)
}

func (t *mapMergeTracer) trace(key string) {
	t.Traces[key] = t.Label
}

func (t *mapMergeTracer) traceNewList(key string, len int) {
	var list []any
	for i := 0; i < len; i++ {
		list = append(list, t.Label)
	}
	t.Traces[key] = list
}

func (t *mapMergeTracer) traceInsertList(key string, len int) {
	var list []any
	for i := 0; i < len; i++ {
		list = append(list, t.Label)
	}
	if existList, ok := t.Traces[key].([]any); ok {
		t.Traces[key] = append(list, existList...)
	} else {
		t.Traces[key] = list
	}
}

func (t *mapMergeTracer) traceAppendList(key string, len int) {
	var list []any
	for i := 0; i < len; i++ {
		list = append(list, t.Label)
	}
	if existList, ok := t.Traces[key].([]any); ok {
		t.Traces[key] = append(existList, list...)
	} else {
		t.Traces[key] = list
	}
}

// endregion
