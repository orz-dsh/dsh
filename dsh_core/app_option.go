package dsh_core

import (
	"github.com/expr-lang/expr/vm"
	"maps"
	"os"
	"os/user"
	"runtime"
	"strings"
)

// region option

const (
	GlobalOptionNameOs       = "_os"
	GlobalOptionNameArch     = "_arch"
	GlobalOptionNameShell    = "_shell"
	GlobalOptionNameHostname = "_hostname"
	GlobalOptionNameUsername = "_username"
)

var globalOptionDefaultOs string
var globalOptionDefaultArch string
var globalOptionDefaultShell string
var globalOptionDefaultHostname string
var globalOptionDefaultUsername string

type appOption struct {
	context        *appContext
	GlobalOptions  map[string]any
	SpecifyOptions map[string]map[string]string
	ProjectOptions map[string]map[string]any
	Assigns        map[string]*appOptionAssign
	Results        map[string]*appOptionResult
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

func loadAppOption(context *appContext, manifest *projectManifest, values map[string]string) (*appOption, error) {
	_os := ""
	if _os = values[GlobalOptionNameOs]; _os == "" {
		_os = getGlobalOptionDefaultOs()
	}
	_arch := ""
	if _arch = values[GlobalOptionNameArch]; _arch == "" {
		_arch = getGlobalOptionDefaultArch()
	}
	_shell := ""
	if _shell = values[GlobalOptionNameShell]; _shell == "" {
		_shell = getGlobalOptionDefaultShell(_os)
	}
	_hostname := ""
	if _hostname = values[GlobalOptionNameHostname]; _hostname == "" {
		hostname, err := getGlobalOptionDefaultHostname()
		if err != nil {
			return nil, err
		}
		_hostname = hostname
	}
	_username := ""
	if _username = values[GlobalOptionNameUsername]; _username == "" {
		username, err := getGlobalOptionDefaultUsername()
		if err != nil {
			return nil, err
		}
		_username = username
	}
	specifyOptions := make(map[string]string)
	for k, v := range values {
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
		context: context,
		GlobalOptions: map[string]any{
			GlobalOptionNameOs:       _os,
			GlobalOptionNameArch:     _arch,
			GlobalOptionNameShell:    _shell,
			GlobalOptionNameHostname: _hostname,
			GlobalOptionNameUsername: _username,
		},
		SpecifyOptions: map[string]map[string]string{
			manifest.Name: specifyOptions,
		},
		ProjectOptions: map[string]map[string]any{},
		Results:        map[string]*appOptionResult{},
		Assigns:        map[string]*appOptionAssign{},
	}
	return option, nil
}

func (o *appOption) mergeGlobalOptions(options map[string]any) map[string]any {
	result := make(map[string]any)
	maps.Copy(result, o.GlobalOptions)
	if options != nil {
		maps.Copy(result, options)
	}
	return result
}

func (o *appOption) loadProjectOptions(manifest *projectManifest) error {
	if _, exist := o.ProjectOptions[manifest.Name]; exist {
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
		options[item.Name] = result.ParsedValue
	}

	verifies := manifest.Option.verifies
	for i := 0; i < len(verifies); i++ {
		verify := verifies[i]
		evaluator := o.context.evaluator.SetRootData("options", o.mergeGlobalOptions(options))
		result, err := evaluator.EvalBoolExpr(verify)
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
	o.ProjectOptions[manifest.Name] = options
	return nil
}

func (o *appOption) getProjectOptions(manifest *projectManifest) map[string]any {
	return o.mergeGlobalOptions(o.ProjectOptions[manifest.Name])
}

func (o *appOption) addAssign(sourceProject string, sourceOption string, assignProject string, assignOption string, assignMapping *vm.Program) error {
	source := sourceProject + "." + sourceOption
	target := assignProject + "." + assignOption
	assign := &appOptionAssign{
		Source:      source,
		FinalSource: source,
		mapping:     assignMapping,
	}
	if sourceAssign, exist := o.Assigns[source]; exist {
		assign.FinalSource = sourceAssign.FinalSource
	}
	if existAssign, exist := o.Assigns[target]; exist {
		if existAssign.FinalSource != assign.FinalSource {
			return errN("add option assign error",
				reason("option assign conflict"),
				kv("target", target),
				kv("assign1", assign.Source),
				kv("assign2", existAssign.Source),
			)
		}
	} else {
		o.Assigns[target] = assign
	}
	return nil
}

func (o *appOption) addResult(projectName string, optionName string, result *appOptionResult) error {
	target := projectName + "." + optionName
	if existResult, exist := o.Results[target]; exist {
		return errN("add option result error",
			reason("option result exists"),
			kv("target", target),
			kv("result1", result),
			kv("result2", existResult),
			kv("result1Assign", result.Assign),
			kv("result2assign", existResult.Assign),
		)
	}
	o.Results[target] = result
	return nil
}

func (o *appOption) findAssignValue(projectName string, optionName string) (*appOptionAssign, *string, error) {
	target := projectName + "." + optionName
	if assign, exist := o.Assigns[target]; exist {
		if result, exist := o.Results[assign.Source]; exist {
			if assign.mapping != nil {
				evaluator := o.context.evaluator.SetRootData("options", o.mergeGlobalOptions(map[string]any{
					"value": result.ParsedValue,
				}))
				mappingResult, err := evaluator.EvalStringExpr(assign.mapping)
				if err != nil {
					return assign, nil, errW(err, "find option assign value error",
						reason("mapping value error"),
						kv("target", target),
						kv("targetAssign", assign),
						kv("targetAssignMapping", assign.mapping.Source().Content()),
						kv("sourceResult", result),
						kv("sourceResultAssign", result.Assign),
					)
				}
				return assign, mappingResult, nil
			} else {
				return assign, &result.RawValue, nil
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

	if specifyOptions, exist := o.SpecifyOptions[manifest.Name]; exist {
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
		RawValue:    rawValue,
		ParsedValue: parsedValue,
		Source:      source,
		Assign:      assign,
	}
	return result, nil
}

func (o *appOption) getGlobalOptionsOs() string {
	return o.GlobalOptions[GlobalOptionNameOs].(string)
}

func (o *appOption) getGlobalOptionsShell() string {
	return o.GlobalOptions[GlobalOptionNameShell].(string)
}

// endregion

// region assign

type appOptionAssign struct {
	Source      string
	FinalSource string
	mapping     *vm.Program
}

func (a *appOptionAssign) DescExtraKeyValues() KVS {
	return KVS{
		kv("mapping", a.mapping.Source().Content()),
	}
}

// endregion

// region result

type appOptionResult struct {
	RawValue    string
	ParsedValue any
	Source      appOptionResultSource
	Assign      *appOptionAssign
}

type appOptionResultSource string

const (
	appOptionResultSourceUnset   = "unset"
	appOptionResultSourceSpecify = "specify"
	appOptionResultSourceAssign  = "assign"
	appOptionResultSourceDefault = "default"
)

// endregion
