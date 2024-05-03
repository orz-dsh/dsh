package dsh_core

import (
	"dsh/dsh_utils"
	"github.com/expr-lang/expr/vm"
	"maps"
	"os"
	"os/user"
	"runtime"
	"strings"
)

const (
	GlobalOptionNameOs                 = "_os"
	GlobalOptionNameArch               = "_arch"
	GlobalOptionNameShell              = "_shell"
	GlobalOptionNameHostname           = "_hostname"
	GlobalOptionNameUsername           = "_username"
	globalOptionNameRuntimeVersion     = "_runtime_version"
	globalOptionNameRuntimeVersionCode = "_runtime_version_code"
)

var globalOptionDefaultOs string
var globalOptionDefaultArch string
var globalOptionDefaultShell string
var globalOptionDefaultHostname string
var globalOptionDefaultUsername string

type appOption struct {
	globalOptions  map[string]any
	specifyOptions map[string]map[string]string
	projectOptions map[string]map[string]any
	assigns        map[string]*appOptionAssign
	results        map[string]*appOptionResult
}

type appOptionAssign struct {
	source      string
	finalSource string
	mapping     *vm.Program
}

type appOptionResult struct {
	rawValue    string
	parsedValue any
	source      appOptionResultSource
	assign      *appOptionAssign
}

type appOptionResultSource string

const (
	appOptionResultSourceUnset   = "unset"
	appOptionResultSourceSpecify = "specify"
	appOptionResultSourceAssign  = "assign"
	appOptionResultSourceDefault = "default"
)

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

func getGlobalOptionDefaultHostname() (string, error) {
	if globalOptionDefaultHostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return "", errW(err, "get global option default hostname error",
				reason("get os hostname error"),
			)
		}
		globalOptionDefaultHostname = hostname
	}
	return globalOptionDefaultHostname, nil
}

func getGlobalOptionDefaultUsername() (string, error) {
	if globalOptionDefaultUsername == "" {
		currentUser, err := user.Current()
		if err != nil {
			return "", errW(err, "get global option default username error",
				reason("get current user error"),
			)
		}
		username := currentUser.Username
		if strings.Contains(username, "\\") {
			username = strings.Split(username, "\\")[1]
		}
		globalOptionDefaultUsername = username
	}
	return globalOptionDefaultUsername, nil
}

func loadAppOption(manifest *projectManifest, options map[string]string) (*appOption, error) {
	_os := ""
	if _os = options[GlobalOptionNameOs]; _os == "" {
		_os = getGlobalOptionDefaultOs()
	}
	_arch := ""
	if _arch = options[GlobalOptionNameArch]; _arch == "" {
		_arch = getGlobalOptionDefaultArch()
	}
	_shell := ""
	if _shell = options[GlobalOptionNameShell]; _shell == "" {
		_shell = getGlobalOptionDefaultShell(_os)
	}
	_hostname := ""
	if _hostname = options[GlobalOptionNameHostname]; _hostname == "" {
		hostname, err := getGlobalOptionDefaultHostname()
		if err != nil {
			return nil, err
		}
		_hostname = hostname
	}
	_username := ""
	if _username = options[GlobalOptionNameUsername]; _username == "" {
		username, err := getGlobalOptionDefaultUsername()
		if err != nil {
			return nil, err
		}
		_username = username
	}
	specifyOptions := make(map[string]string)
	for k, v := range options {
		if strings.HasPrefix(k, "_") {
			// global option
			continue
		}
		if strings.Contains(k, ".") {
			// Because of the Option Assign feature
			// You should not specify option values for items other than the main project
			continue
		}
		specifyOptions[k] = v
	}
	option := &appOption{
		globalOptions: map[string]any{
			GlobalOptionNameOs:                 _os,
			GlobalOptionNameArch:               _arch,
			GlobalOptionNameShell:              _shell,
			GlobalOptionNameHostname:           _hostname,
			GlobalOptionNameUsername:           _username,
			globalOptionNameRuntimeVersion:     string(dsh_utils.GetRuntimeVersion()),
			globalOptionNameRuntimeVersionCode: dsh_utils.GetRuntimeVersionCode(),
		},
		specifyOptions: map[string]map[string]string{
			manifest.Name: specifyOptions,
		},
		projectOptions: make(map[string]map[string]any),
		results:        make(map[string]*appOptionResult),
		assigns:        make(map[string]*appOptionAssign),
	}
	return option, nil
}

func (o *appOption) mergeGlobalOptions(options map[string]any) map[string]any {
	result := make(map[string]any)
	maps.Copy(result, o.globalOptions)
	if options != nil {
		maps.Copy(result, options)
	}
	return result
}

func (o *appOption) loadProjectOptions(manifest *projectManifest) error {
	if _, exist := o.projectOptions[manifest.Name]; exist {
		return nil
	}
	options := make(map[string]any)
	for i := 0; i < len(manifest.Option.Items); i++ {
		item := manifest.Option.Items[i]
		result, err := o.findResult(manifest, item)
		if err != nil {
			return errW(err, "load project options error",
				reason("find option result error"),
				kv("projectName", manifest.Name),
				kv("projectPath", manifest.projectPath),
				kv("optionName", item.Name),
			)
		}
		if err = o.addResult(manifest.Name, item.Name, result); err != nil {
			return errW(err, "load project options error",
				reason("add option result error"),
				kv("projectName", manifest.Name),
				kv("projectPath", manifest.projectPath),
				kv("optionName", item.Name),
			)
		}
		options[item.Name] = result.parsedValue
	}

	verifies := manifest.Option.verifies
	for i := 0; i < len(verifies); i++ {
		verify := verifies[i]
		result, err := dsh_utils.EvalExprReturnBool(verify, o.mergeGlobalOptions(options))
		if err != nil {
			return errW(err, "load project options error",
				reason("eval verify error"),
				kv("projectName", manifest.Name),
				kv("projectPath", manifest.projectPath),
				kv("verify", verify.Source().Content()),
			)
		}
		if !result {
			return errN("load project options error",
				reason("verify options error"),
				kv("projectName", manifest.Name),
				kv("projectPath", manifest.projectPath),
				kv("verify", verify.Source().Content()),
			)
		}
	}

	for i := 0; i < len(manifest.Option.Items); i++ {
		item := manifest.Option.Items[i]
		for j := 0; j < len(item.Assigns); j++ {
			assign := item.Assigns[j]
			if err := o.addAssign(manifest.Name, item.Name, assign.Project, assign.Option, assign.mapping); err != nil {
				return errW(err, "load project options error",
					reason("add option assign error"),
					kv("projectName", manifest.Name),
					kv("projectPath", manifest.projectPath),
					kv("optionName", item.Name),
					kv("assignProject", assign.Project),
					kv("assignOption", assign.Option),
				)
			}
		}
	}
	o.projectOptions[manifest.Name] = options
	return nil
}

func (o *appOption) getProjectOptions(manifest *projectManifest) map[string]any {
	return o.mergeGlobalOptions(o.projectOptions[manifest.Name])
}

func (o *appOption) evalProjectMatchExpr(manifest *projectManifest, expr *vm.Program) (bool, error) {
	matched, err := dsh_utils.EvalExprReturnBool(expr, o.getProjectOptions(manifest))
	if err != nil {
		return false, errW(err, "eval project match expr error",
			reason("eval expr error"),
			kv("projectName", manifest.Name),
			kv("projectPath", manifest.projectPath),
			kv("matchExpr", expr.Source().Content()),
		)
	}
	return matched, nil
}

func (o *appOption) addAssign(sourceProject string, sourceOption string, assignProject string, assignOption string, assignMapping *vm.Program) error {
	source := sourceProject + "." + sourceOption
	target := assignProject + "." + assignOption
	assign := &appOptionAssign{
		source:      source,
		finalSource: source,
		mapping:     assignMapping,
	}
	if sourceAssign, exist := o.assigns[source]; exist {
		assign.finalSource = sourceAssign.finalSource
	}
	if existAssign, exist := o.assigns[target]; exist {
		if existAssign.finalSource != assign.finalSource {
			return errN("add option assign error",
				reason("option assign conflict"),
				kv("target", target),
				kv("assign1", assign.source),
				kv("assign2", existAssign.source),
			)
		}
	} else {
		o.assigns[target] = assign
	}
	return nil
}

func (o *appOption) addResult(projectName string, optionName string, result *appOptionResult) error {
	target := projectName + "." + optionName
	if existResult, exist := o.results[target]; exist {
		return errN("add option result error",
			reason("option result exists"),
			kv("target", target),
			kv("result1", result),
			kv("result2", existResult),
			kv("result1Assign", result.assign),
			kv("result2assign", existResult.assign),
		)
	}
	o.results[target] = result
	return nil
}

func (o *appOption) findAssignValue(projectName string, optionName string) (*appOptionAssign, *string, error) {
	target := projectName + "." + optionName
	if assign, exist := o.assigns[target]; exist {
		if result, exist := o.results[assign.source]; exist {
			if assign.mapping != nil {
				mappingResult, err := dsh_utils.EvalExprReturnString(assign.mapping, o.mergeGlobalOptions(map[string]any{
					"value": result.parsedValue,
				}))
				if err != nil {
					return assign, nil, errW(err, "find option assign value error",
						reason("mapping value error"),
						kv("target", target),
						kv("targetAssign", assign),
						kv("targetAssignMapping", assign.mapping.Source().Content()),
						kv("sourceResult", result),
						kv("sourceResultAssign", result.assign),
					)
				}
				return assign, mappingResult, nil
			} else {
				return assign, &result.rawValue, nil
			}
		} else {
			return assign, nil, errN("find option assign value error",
				reason("source result not found"),
				kv("target", target),
				kv("targetAssign", assign),
			)
		}
	}
	return nil, nil, nil
}

func (o *appOption) findResult(manifest *projectManifest, item *projectManifestOptionItem) (result *appOptionResult, err error) {
	found := false
	var rawValue string
	var parsedValue any = nil
	var source appOptionResultSource = appOptionResultSourceUnset
	var assign *appOptionAssign = nil

	if specifyOptions, exist := o.specifyOptions[manifest.Name]; exist {
		if value, exist := specifyOptions[item.Name]; exist {
			rawValue = value
			parsedValue, err = item.parseValue(rawValue)
			if err != nil {
				return nil, errW(err, "find option result error",
					reason("parse specify value error"),
					kv("projectName", manifest.Name),
					kv("projectPath", manifest.projectPath),
					kv("optionName", item.Name),
					kv("optionValue", rawValue),
				)
			}
			source = appOptionResultSourceSpecify
			found = true
		}
	}

	if !found {
		var assignValue *string
		assign, assignValue, err = o.findAssignValue(manifest.Name, item.Name)
		if err != nil {
			return nil, errW(err, "find option result error",
				reason("get assign value error"),
				kv("projectName", manifest.Name),
				kv("projectPath", manifest.projectPath),
				kv("optionName", item.Name),
			)
		}
		if assign != nil {
			if assignValue != nil {
				rawValue = *assignValue
				parsedValue, err = item.parseValue(rawValue)
				if err != nil {
					return nil, errW(err, "find option result error",
						reason("parse assign value error"),
						kv("projectName", manifest.Name),
						kv("projectPath", manifest.projectPath),
						kv("optionName", item.Name),
						kv("optionValue", rawValue),
					)
				}
			}
			source = appOptionResultSourceAssign
			found = true
		}
	}

	if !found {
		if item.defaultParsedValue != nil {
			rawValue = item.defaultRawValue
			parsedValue = item.defaultParsedValue
			source = appOptionResultSourceDefault
			found = true
		}
	}

	if parsedValue == nil && !item.Optional {
		return nil, errN("find option result error",
			reason("option value empty"),
			kv("projectName", manifest.Name),
			kv("projectPath", manifest.projectPath),
			kv("optionName", item.Name),
			kv("source", source),
			kv("assign", assign),
		)
	}

	result = &appOptionResult{
		rawValue:    rawValue,
		parsedValue: parsedValue,
		source:      source,
		assign:      assign,
	}
	return result, nil
}
