package core

import (
	"github.com/orz-dsh/dsh/utils"
	"os"
	"strings"
)

type Global struct {
	logger     *Logger
	systemInfo *SystemInfo
	variables  map[string]any
}

func MakeGlobal(logger *Logger, variables map[string]string) (*Global, error) {
	systemInfo, err := utils.GetSystemInfo()
	if err != nil {
		return nil, errW(err, "make global error",
			reason("get system info error"),
		)
	}
	global := &Global{
		logger:     logger,
		systemInfo: systemInfo,
		variables:  mergeGlobalVariables(variables),
	}
	return global, nil
}

func mergeGlobalVariables(variables map[string]string) map[string]any {
	result := map[string]any{}
	for _, e := range os.Environ() {
		equalIndex := strings.Index(e, "=")
		key := e[:equalIndex]
		if name, found := strings.CutPrefix(key, "DSH_GLOBAL_"); found {
			name = strings.ReplaceAll(strings.ToLower(name), "-", "_")
			result[name] = e[equalIndex+1:]
		}
	}
	for k, v := range variables {
		result[k] = v
	}
	return result
}

func (g *Global) DescExtraKeyValues() KVS {
	return KVS{
		kv("systemInfo", g.systemInfo),
		kv("variables", g.variables),
	}
}
