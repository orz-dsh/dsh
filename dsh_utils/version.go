package dsh_utils

import (
	"strconv"
	"strings"
)

type Version string

const runtimeVersion Version = "1.0.0"

var runtimeVersionCode int32

func init() {
	var err error
	runtimeVersionCode, err = runtimeVersion.GetVersionCode()
	if err != nil {
		panic(err)
	}
}

func GetRuntimeVersionCode() int32 {
	return runtimeVersionCode
}

func GetRuntimeVersionName() string {
	return string(runtimeVersion)
}

func CheckRuntimeVersion(minVersion Version, maxVersion Version) (err error) {
	minVersionCode := int32(0)
	if minVersion != "" {
		minVersionCode, err = minVersion.GetVersionCode()
		if err != nil {
			return WrapError(err, "min version get code failed", map[string]any{
				"minVersion": minVersion,
			})
		}
	}
	maxVersionCode := int32(999999999)
	if maxVersion != "" {
		maxVersionCode, err = maxVersion.GetVersionCode()
		if err != nil {
			return WrapError(err, "max version get code failed", map[string]any{
				"maxVersion": maxVersion,
			})
		}
	}
	if runtimeVersionCode >= minVersionCode && runtimeVersionCode <= maxVersionCode {
		return nil
	}
	return NewError("runtime version check failed", map[string]any{
		"runtimeVersion": runtimeVersion,
		"minVersion":     minVersion,
		"maxVersion":     maxVersion,
	})
}

func (v Version) GetVersionCode() (versionCode int32, err error) {
	versionStr := string(v)
	fragmentStr := strings.Split(versionStr, ".")
	if len(fragmentStr) < 1 || len(fragmentStr) > 3 {
		return 0, NewError("version format invalid", map[string]any{
			"version": versionStr,
		})
	}
	var fragmentCode []int32
	for i := 0; i < len(fragmentStr); i++ {
		code, err := strconv.Atoi(fragmentStr[i])
		if err != nil {
			return 0, NewError("version format invalid", map[string]any{
				"version": versionStr,
			})
		}
		if code > 999 {
			return 0, NewError("version format invalid", map[string]any{
				"version": versionStr,
			})
		}
		fragmentCode = append(fragmentCode, int32(code))
	}
	if len(fragmentCode) == 1 {
		return fragmentCode[0] * 1000000, nil
	}
	if len(fragmentCode) == 2 {
		return fragmentCode[0]*1000000 + fragmentCode[1]*1000, nil
	}
	return fragmentCode[0]*1000000 + fragmentCode[1]*1000 + fragmentCode[2], nil
}
