package utils

import (
	"encoding/json"
	"reflect"
	"strconv"
)

type CastType string

const (
	CastTypeString  CastType = "string"
	CastTypeBool    CastType = "bool"
	CastTypeInteger CastType = "integer"
	CastTypeDecimal CastType = "decimal"
	CastTypeObject  CastType = "object"
	CastTypeArray   CastType = "array"
)

func CastToString(value any) (*string, error) {
	if value != nil {
		var stringValue string
		switch value.(type) {
		case bool:
			stringValue = strconv.FormatBool(value.(bool))
		case string:
			stringValue = value.(string)
		case int:
			stringValue = strconv.Itoa(value.(int))
		case int8:
			stringValue = strconv.Itoa(int(value.(int8)))
		case int16:
			stringValue = strconv.Itoa(int(value.(int16)))
		case int32:
			stringValue = strconv.Itoa(int(value.(int32)))
		case int64:
			stringValue = strconv.FormatInt(value.(int64), 10)
		case uint:
			stringValue = strconv.FormatUint(uint64(value.(uint)), 10)
		case uint8:
			stringValue = strconv.FormatUint(uint64(value.(uint8)), 10)
		case uint16:
			stringValue = strconv.FormatUint(uint64(value.(uint16)), 10)
		case uint32:
			stringValue = strconv.FormatUint(uint64(value.(uint32)), 10)
		case uint64:
			stringValue = strconv.FormatUint(value.(uint64), 10)
		case float32:
			stringValue = strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32)
		case float64:
			stringValue = strconv.FormatFloat(value.(float64), 'f', -1, 64)
		case []any:
			if bytes, err := json.Marshal(value.([]any)); err != nil {
				return nil, ErrW(err, "cast to string error",
					Reason("marshal array value json error"),
					KV("value", value),
				)
			} else {
				stringValue = string(bytes)
			}
		case map[string]any:
			if bytes, err := json.Marshal(value.(map[string]any)); err != nil {
				return nil, ErrW(err, "cast to string error",
					Reason("marshal map value json error"),
					KV("value", value),
				)
			} else {
				stringValue = string(bytes)
			}
		default:
			return nil, ErrN("cast to string error",
				Reason("unsupported value type"),
				KV("value", value),
				KV("type", reflect.TypeOf(value)),
			)
		}
		return &stringValue, nil
	}
	return nil, nil
}

func CastToBool(value any) (*bool, error) {
	if value != nil {
		var boolValue bool
		switch value.(type) {
		case bool:
			boolValue = value.(bool)
		case string:
			boolValue = value.(string) != ""
		case int:
			boolValue = value.(int) != 0
		case int8:
			boolValue = value.(int8) != 0
		case int16:
			boolValue = value.(int16) != 0
		case int32:
			boolValue = value.(int32) != 0
		case int64:
			boolValue = value.(int64) != 0
		case uint:
			boolValue = value.(uint) != 0
		case uint8:
			boolValue = value.(uint8) != 0
		case uint16:
			boolValue = value.(uint16) != 0
		case uint32:
			boolValue = value.(uint32) != 0
		case uint64:
			boolValue = value.(uint64) != 0
		case float32:
			boolValue = value.(float32) != 0
		case float64:
			boolValue = value.(float64) != 0
		case []any:
			boolValue = len(value.([]any)) > 0
		case map[string]any:
			boolValue = len(value.(map[string]any)) > 0
		default:
			return nil, ErrN("cast to bool error",
				Reason("unsupported value type"),
				KV("value", value),
				KV("type", reflect.TypeOf(value)),
			)
		}
		return &boolValue, nil
	}
	return nil, nil
}

func CastToInteger(value any) (*int64, error) {
	if value != nil {
		var integerValue int64
		var err error
		switch value.(type) {
		case bool:
			if value.(bool) {
				integerValue = 1
			} else {
				integerValue = 0
			}
		case string:
			integerValue, err = strconv.ParseInt(value.(string), 10, 64)
			if err != nil {
				return nil, ErrW(err, "cast to integer error",
					Reason("parse int error"),
					KV("value", value),
				)
			}
		case int:
			integerValue = int64(value.(int))
		case int8:
			integerValue = int64(value.(int8))
		case int16:
			integerValue = int64(value.(int16))
		case int32:
			integerValue = int64(value.(int32))
		case int64:
			integerValue = value.(int64)
		case uint:
			integerValue = int64(value.(uint))
		case uint8:
			integerValue = int64(value.(uint8))
		case uint16:
			integerValue = int64(value.(uint16))
		case uint32:
			integerValue = int64(value.(uint32))
		case uint64:
			integerValue = int64(value.(uint64))
		case float32:
			integerValue = int64(value.(float32))
		case float64:
			integerValue = int64(value.(float64))
		default:
			return nil, ErrN("cast to integer error",
				Reason("unsupported value type"),
				KV("value", value),
				KV("type", reflect.TypeOf(value)),
			)
		}
		return &integerValue, nil
	}
	return nil, nil
}

func CastToDecimal(value any) (*float64, error) {
	if value != nil {
		var decimalValue float64
		var err error
		switch value.(type) {
		case bool:
			if value.(bool) {
				decimalValue = 1
			} else {
				decimalValue = 0
			}
		case string:
			decimalValue, err = strconv.ParseFloat(value.(string), 64)
			if err != nil {
				return nil, ErrW(err, "cast to decimal error",
					Reason("parse float error"),
					KV("value", value),
				)
			}
		case int:
			decimalValue = float64(value.(int))
		case int8:
			decimalValue = float64(value.(int8))
		case int16:
			decimalValue = float64(value.(int16))
		case int32:
			decimalValue = float64(value.(int32))
		case int64:
			decimalValue = float64(value.(int64))
		case uint:
			decimalValue = float64(value.(uint))
		case uint8:
			decimalValue = float64(value.(uint8))
		case uint16:
			decimalValue = float64(value.(uint16))
		case uint32:
			decimalValue = float64(value.(uint32))
		case uint64:
			decimalValue = float64(value.(uint64))
		case float32:
			decimalValue = float64(value.(float32))
		case float64:
			decimalValue = value.(float64)
		default:
			return nil, ErrN("cast to decimal error",
				Reason("unsupported value type"),
				KV("value", value),
				KV("type", reflect.TypeOf(value)),
			)
		}
		return &decimalValue, nil
	}
	return nil, nil
}

func CastToObject(value any) (map[string]any, error) {
	if value != nil {
		var objectValue map[string]any
		switch value.(type) {
		case map[string]any:
			objectValue = value.(map[string]any)
		case string:
			if err := json.Unmarshal([]byte(value.(string)), &objectValue); err != nil {
				return nil, ErrW(err, "cast to object error",
					Reason("unmarshal json error"),
					KV("value", value),
				)
			}
		default:
			return nil, ErrN("cast to object error",
				Reason("unsupported value type"),
				KV("value", value),
				KV("type", reflect.TypeOf(value)),
			)
		}
		return objectValue, nil
	}
	return nil, nil
}

func CastToArray(value any) ([]any, error) {
	if value != nil {
		var arrayValue []any
		switch value.(type) {
		case []any:
			arrayValue = value.([]any)
		case string:
			if err := json.Unmarshal([]byte(value.(string)), &arrayValue); err != nil {
				return nil, ErrW(err, "cast to array error",
					Reason("unmarshal json error"),
					KV("value", value),
				)
			}
		default:
			return nil, ErrN("cast to array error",
				Reason("unsupported value type"),
				KV("value", value),
				KV("type", reflect.TypeOf(value)),
			)
		}
		return arrayValue, nil
	}
	return nil, nil
}

func Cast(value any, typ CastType) (any, error) {
	switch typ {
	case CastTypeString:
		if stringValue, err := CastToString(value); err != nil {
			return nil, err
		} else if stringValue != nil {
			return *stringValue, nil
		}
		return nil, nil
	case CastTypeBool:
		if boolValue, err := CastToBool(value); err != nil {
			return nil, err
		} else if boolValue != nil {
			return *boolValue, nil
		}
		return nil, nil
	case CastTypeInteger:
		if integerValue, err := CastToInteger(value); err != nil {
			return nil, err
		} else if integerValue != nil {
			return *integerValue, nil
		}
		return nil, nil
	case CastTypeDecimal:
		if decimalValue, err := CastToDecimal(value); err != nil {
			return nil, err
		} else if decimalValue != nil {
			return *decimalValue, nil
		}
		return nil, nil
	case CastTypeObject:
		if objectValue, err := CastToObject(value); err != nil {
			return nil, err
		} else if objectValue != nil {
			return objectValue, nil
		}
		return nil, nil
	case CastTypeArray:
		if arrayValue, err := CastToArray(value); err != nil {
			return nil, err
		} else if arrayValue != nil {
			return arrayValue, nil
		}
		return nil, nil
	default:
		Impossible()
	}
	return nil, nil
}

func CastSlice[T any](slice []T, typ CastType) ([]any, error) {
	result := make([]any, 0, len(slice))
	for i := 0; i < len(slice); i++ {
		item, err := Cast(slice[i], typ)
		if err != nil {
			return nil, ErrW(err, "cast slice error",
				Reason("cast item error"),
				KV("slice", slice),
				KV("index", i),
			)
		}
		result = append(result, item)
	}
	return result, nil
}
