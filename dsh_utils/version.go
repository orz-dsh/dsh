package dsh_utils

import (
	"strconv"
	"strings"
)

type Version string

const runtimeVersion Version = "1.0.0"

var runtimeVersionNumber int32

func init() {
	var err error
	runtimeVersionNumber, err = runtimeVersion.GetVersionNumber()
	if err != nil {
		panic(err)
	}
}

func (v Version) GetVersionNumber() (versionNumber int32, err error) {
	versionStr := string(v)
	fragmentStr := strings.Split(versionStr, ".")
	if len(fragmentStr) < 1 || len(fragmentStr) > 3 {
		return 0, NewError("version format invalid", map[string]any{
			"version": versionStr,
		})
	}
	var fragmentNumber []int32
	for i := 0; i < len(fragmentStr); i++ {
		number, err := strconv.Atoi(fragmentStr[i])
		if err != nil {
			return 0, NewError("version format invalid", map[string]any{
				"version": versionStr,
			})
		}
		if number > 999 {
			return 0, NewError("version format invalid", map[string]any{
				"version": versionStr,
			})
		}
		fragmentNumber = append(fragmentNumber, int32(number))
	}
	if len(fragmentNumber) == 1 {
		return fragmentNumber[0] * 1000000, nil
	}
	if len(fragmentNumber) == 2 {
		return fragmentNumber[0]*1000000 + fragmentNumber[1]*1000, nil
	}
	return fragmentNumber[0]*1000000 + fragmentNumber[1]*1000 + fragmentNumber[2], nil
}

func CheckRuntimeVersion(minVersion Version, maxVersion Version) (err error) {
	minVersionNumber := int32(0)
	if minVersion != "" {
		minVersionNumber, err = minVersion.GetVersionNumber()
		if err != nil {
			return WrapError(err, "min version get number failed", map[string]any{
				"minVersion": minVersion,
			})
		}
	}
	maxVersionNumber := int32(999999999)
	if maxVersion != "" {
		maxVersionNumber, err = maxVersion.GetVersionNumber()
		if err != nil {
			return WrapError(err, "max version get number failed", map[string]any{
				"maxVersion": maxVersion,
			})
		}
	}
	if runtimeVersionNumber >= minVersionNumber && runtimeVersionNumber <= maxVersionNumber {
		return nil
	}
	return NewError("runtime version check failed", map[string]any{
		"runtimeVersion": runtimeVersion,
		"minVersion":     minVersion,
		"maxVersion":     maxVersion,
	})
}
