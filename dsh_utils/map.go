package dsh_utils

func MapStrStrToStrAny(m map[string]string) map[string]any {
	result := map[string]any{}
	for k, v := range m {
		result[k] = v
	}
	return result
}
