package dsh_utils

import "reflect"

type MapMergeMode string

const (
	MapMergeModeReplace MapMergeMode = "replace"
	MapMergeModeInsert  MapMergeMode = "insert"
	MapMergeModeRootKey string       = "$root"
)

func MapStrStrToStrAny(m map[string]string) map[string]any {
	result := map[string]any{}
	for k, v := range m {
		result[k] = v
	}
	return result
}

func MergeMap(target map[string]any, source map[string]any, modes map[string]MapMergeMode) (map[string]any, error) {
	if modes == nil {
		modes = map[string]MapMergeMode{}
	}
	if target != nil && modes[MapMergeModeRootKey] == MapMergeModeReplace {
		clear(target)
	}
	return mergeMap(target, source, modes, "")
}

func mergeMap(target map[string]any, source map[string]any, modes map[string]MapMergeMode, parentKey string) (_ map[string]any, err error) {
	if target == nil {
		target = map[string]any{}
	}
	for k, v := range source {
		switch v.(type) {
		case map[string]any:
			key := k
			if parentKey != "" {
				key = parentKey + "." + k
			}
			sourceMap := v.(map[string]any)
			targetValue := target[k]
			if targetValue == nil {
				if target[k], err = mergeMap(nil, sourceMap, modes, key); err != nil {
					return nil, err
				}
			} else if targetMap, ok := targetValue.(map[string]any); ok {
				if mode, exist := modes[key]; exist {
					switch mode {
					case MapMergeModeReplace:
						if target[k], err = mergeMap(nil, sourceMap, modes, key); err != nil {
							return nil, err
						}
					default:
						return nil, errN("merge map error",
							reason("merge mode invalid"),
							kv("key", key),
							kv("mergeMode", mode),
							kv("supportModes", []MapMergeMode{
								MapMergeModeReplace,
							}),
						)
					}
				} else {
					if target[k], err = mergeMap(targetMap, sourceMap, modes, key); err != nil {
						return nil, err
					}
				}
			} else {
				return nil, errN("merge map error",
					reason("source type not match target type"),
					kv("key", key),
					kv("sourceType", reflect.TypeOf(sourceMap)),
					kv("targetType", reflect.TypeOf(targetValue)),
				)
			}
		case []any:
			sourceKey := k
			if parentKey != "" {
				sourceKey = parentKey + "." + k
			}
			sourceList := v.([]any)
			targetValue := target[k]
			if targetValue == nil {
				target[k] = sourceList
			} else if targetList, ok := targetValue.([]any); ok {
				if mode, exist := modes[sourceKey]; exist {
					if mode == MapMergeModeReplace {
						target[k] = sourceList
					} else if mode == MapMergeModeInsert {
						target[k] = append(sourceList, targetList...)
					} else {
						return nil, errN("merge map error",
							reason("merge type invalid"),
							kv("key", sourceKey),
							kv("mergeMode", mode),
							kv("supportMode", []MapMergeMode{
								MapMergeModeReplace,
								MapMergeModeInsert,
							}),
						)
					}
				} else {
					target[k] = append(targetList, sourceList...)
				}
			} else {
				return nil, errN("merge map error",
					reason("source type not match target type"),
					kv("key", sourceKey),
					kv("sourceType", reflect.TypeOf(sourceList)),
					kv("targetType", reflect.TypeOf(targetValue)),
				)
			}
		default:
			target[k] = v
		}
	}
	return target, nil
}
