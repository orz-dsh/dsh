package utils

import (
	"strconv"
	"strings"
)

type Version string

const _runtimeVersion Version = "1.0.0"

var _runtimeVersionCode int32

func init() {
	var err error
	_runtimeVersionCode, err = _runtimeVersion.GetVersionCode()
	if err != nil {
		panic(err)
	}
}

func GetRuntimeVersion() Version {
	return _runtimeVersion
}

func GetRuntimeVersionCode() int32 {
	return _runtimeVersionCode
}

func CheckRuntimeVersion(minVersion Version, maxVersion Version) (err error) {
	minVersionCode := int32(0)
	if minVersion != "" {
		minVersionCode, err = minVersion.GetVersionCode()
		if err != nil {
			return ErrW(err, "check runtime version error",
				Reason("get min version code error"),
				KV("minVersion", minVersion),
			)
		}
	}
	maxVersionCode := int32(999999999)
	if maxVersion != "" {
		maxVersionCode, err = maxVersion.GetVersionCode()
		if err != nil {
			return ErrW(err, "check runtime version error",
				Reason("get max version code error"),
				KV("maxVersion", maxVersion),
			)
		}
	}
	if _runtimeVersionCode >= minVersionCode && _runtimeVersionCode <= maxVersionCode {
		return nil
	}
	return ErrN("check runtime version error",
		Reason("runtime version incompatible"),
		KV("runtimeVersion", _runtimeVersion),
		KV("minVersion", minVersion),
		KV("maxVersion", maxVersion),
	)
}

func (v Version) GetVersionCode() (versionCode int32, err error) {
	versionStr := string(v)
	fragmentStr := strings.Split(versionStr, ".")
	if len(fragmentStr) < 1 || len(fragmentStr) > 3 {
		return 0, ErrN("get version code error",
			Reason("format error"),
			KV("version", versionStr),
		)
	}
	var fragmentCode []int32
	for i := 0; i < len(fragmentStr); i++ {
		code, err := strconv.Atoi(fragmentStr[i])
		if err != nil {
			return 0, ErrN("get version code error",
				Reason("format error"),
				KV("version", versionStr),
			)
		}
		if code > 999 {
			return 0, ErrN("get version code error",
				Reason("format error"),
				KV("version", versionStr),
			)
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
