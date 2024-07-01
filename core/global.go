package core

import (
	. "github.com/orz-dsh/dsh/utils"
	"os"
	"regexp"
	"strings"
)

var globalVariableNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*[a-z0-9]$")

type Global struct {
	logger     *Logger
	systemInfo *SystemInfo
	variables  map[string]any
}

func MakeGlobal(logger *Logger, variables map[string]string) (*Global, error) {
	systemInfo, err := GetSystemInfo()
	if err != nil {
		return nil, ErrW(err, "make global error",
			Reason("get system info error"),
		)
	}
	finalVariables, err := mergeGlobalVariables(variables)
	if err != nil {
		return nil, ErrW(err, "make global error",
			Reason("merge global variables error"),
		)
	}
	global := &Global{
		logger:     logger,
		systemInfo: systemInfo,
		variables:  finalVariables,
	}
	return global, nil
}

func mergeGlobalVariables(variables map[string]string) (map[string]any, error) {
	result := map[string]any{}
	for _, e := range os.Environ() {
		equalIndex := strings.Index(e, "=")
		key := e[:equalIndex]
		if name, found := strings.CutPrefix(key, "DSH_GLOBAL_"); found {
			name = strings.ReplaceAll(strings.ToLower(name), "-", "_")
			if !globalVariableNameCheckRegex.MatchString(name) {
				return nil, ErrN("merge global variables error",
					Reason("invalid variable name"),
					KV("name", name),
					KV("env", key),
				)
			}
			result[name] = e[equalIndex+1:]
		}
	}
	for k, v := range variables {
		if !globalVariableNameCheckRegex.MatchString(k) {
			return nil, ErrN("merge global variables error",
				Reason("invalid variable name"),
				KV("name", k),
			)
		}
		result[k] = v
	}
	return result, nil
}

func (g *Global) DescExtraKeyValues() KVS {
	return KVS{
		KV("systemInfo", g.systemInfo),
		KV("variables", g.variables),
	}
}
