package dsh_core

import (
	"dsh/dsh_utils"
	"maps"
	"os"
	"os/user"
	"runtime"
	"strings"
)

var globalOptionDefaultOs string
var globalOptionDefaultArch string
var globalOptionDefaultShell string
var globalOptionDefaultHostname string
var globalOptionDefaultUsername string

type globalOption struct {
	items map[string]any
}

func newGlobalOption(optionValues map[string]string, logger *dsh_utils.Logger) *globalOption {
	_os := ""
	if _os = optionValues["_os"]; _os == "" {
		_os = getGlobalOptionDefaultOs()
	}
	_arch := ""
	if _arch = optionValues["_arch"]; _arch == "" {
		_arch = getGlobalOptionDefaultArch()
	}
	_shell := ""
	if _shell = optionValues["_shell"]; _shell == "" {
		_shell = getGlobalOptionDefaultShell(_os)
	}
	_hostname := ""
	if _hostname = optionValues["_hostname"]; _hostname == "" {
		_hostname = getGlobalOptionDefaultHostname(logger)
	}
	_username := ""
	if _username = optionValues["_username"]; _username == "" {
		_username = getGlobalOptionDefaultUsername(logger)
	}
	return &globalOption{
		items: map[string]any{
			"_os":                   _os,
			"_arch":                 _arch,
			"_shell":                _shell,
			"_hostname":             _hostname,
			"_username":             _username,
			"_runtime_version_name": dsh_utils.GetRuntimeVersionName(),
			"_runtime_version_code": dsh_utils.GetRuntimeVersionCode(),
		},
	}
}

func getGlobalOptionDefaultOs() string {
	if globalOptionDefaultOs == "" {
		globalOptionDefaultOs = strings.ToLower(runtime.GOOS)
	}
	return globalOptionDefaultOs
}

func getGlobalOptionDefaultArch() string {
	if globalOptionDefaultArch == "" {
		arch := strings.ToLower(runtime.GOARCH)
		if arch == "amd64" {
			globalOptionDefaultArch = "x64"
		} else if arch == "386" {
			globalOptionDefaultArch = "x32"
		} else {
			globalOptionDefaultArch = arch
		}
	}
	return globalOptionDefaultArch
}

func getGlobalOptionDefaultShell(os string) string {
	if globalOptionDefaultShell == "" {
		if os == "windows" {
			globalOptionDefaultShell = "cmd"
		} else {
			globalOptionDefaultShell = "sh"
		}
	}
	return globalOptionDefaultShell
}

func getGlobalOptionDefaultHostname(logger *dsh_utils.Logger) string {
	if globalOptionDefaultHostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			logger.Panic("%+v", dsh_utils.WrapError(err, "hostname get failed", nil))
		} else {
			globalOptionDefaultHostname = hostname
		}
	}
	return globalOptionDefaultHostname
}

func getGlobalOptionDefaultUsername(logger *dsh_utils.Logger) string {
	if globalOptionDefaultUsername == "" {
		currentUser, err := user.Current()
		if err != nil {
			logger.Panic("%+v", dsh_utils.WrapError(err, "current user get failed", nil))
		} else {
			username := currentUser.Username
			if strings.Contains(username, "\\") {
				username = strings.Split(username, "\\")[1]
			}
			globalOptionDefaultUsername = username
		}
	}
	return globalOptionDefaultUsername
}

func (option *globalOption) mergeItems(items map[string]any) map[string]any {
	result := make(map[string]any)
	maps.Copy(result, option.items)
	maps.Copy(result, items)
	return result
}
