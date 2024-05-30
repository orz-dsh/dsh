package dsh_core

import (
	"github.com/expr-lang/expr/vm"
	"maps"
	"strings"
)

// region option

type appOption struct {
	evaluator    *Evaluator
	GenericItems appOptionGenericItems
	SpecifyItems appOptionSpecifyItems
	Assigns      map[string]*appOptionAssign
	Results      map[string]*appOptionResult
}

func newAppOption(systemInfo *SystemInfo, evaluator *Evaluator, projectName string, specifyItems map[string]string) *appOption {
	return &appOption{
		evaluator:    evaluator,
		GenericItems: newAppOptionGenericItems(systemInfo, specifyItems),
		SpecifyItems: newAppOptionSpecifyItems(projectName, specifyItems),
		Results:      map[string]*appOptionResult{},
		Assigns:      map[string]*appOptionAssign{},
	}
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
				kv("assign", assign),
				kv("existAssign", existAssign),
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
			kv("result", result),
			kv("existResult", existResult),
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
				evaluator := o.evaluator.SetRootData("options", o.GenericItems.merge(map[string]any{
					"value": result.ParsedValue,
				}))
				mappingResult, err := evaluator.EvalStringExpr(assign.mapping)
				if err != nil {
					return assign, nil, errW(err, "find option assign value error",
						reason("mapping value error"),
						kv("target", target),
						kv("assign", assign),
						kv("result", result),
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
				kv("assign", assign),
			)
		}
	}
	return nil, nil, nil
}

func (o *appOption) findResult(manifest *ProjectManifest, declare *projectOptionDeclareEntity) (result *appOptionResult, err error) {
	found := false
	var rawValue string
	var parsedValue any = nil
	var source appOptionResultSource = appOptionResultSourceUnset
	var assign *appOptionAssign = nil

	if specifyItems, exist := o.SpecifyItems[manifest.projectName]; exist {
		if value, exist := specifyItems[declare.Name]; exist {
			rawValue = value
			parsedValue, err = declare.parseValue(rawValue)
			if err != nil {
				return nil, errW(err, "find option result error",
					reason("parse specify value error"),
					kv("projectName", manifest.projectName),
					kv("projectPath", manifest.projectPath),
					kv("optionName", declare.Name),
					kv("optionValue", rawValue),
				)
			}
			source = appOptionResultSourceSpecify
			found = true
		}
	}

	if !found {
		var assignValue *string
		assign, assignValue, err = o.findAssignValue(manifest.projectName, declare.Name)
		if err != nil {
			return nil, errW(err, "find option result error",
				reason("get assign value error"),
				kv("projectName", manifest.projectName),
				kv("projectPath", manifest.projectPath),
				kv("optionName", declare.Name),
			)
		}
		if assign != nil {
			if assignValue != nil {
				rawValue = *assignValue
				parsedValue, err = declare.parseValue(rawValue)
				if err != nil {
					return nil, errW(err, "find option result error",
						reason("parse assign value error"),
						kv("projectName", manifest.projectName),
						kv("projectPath", manifest.projectPath),
						kv("optionName", declare.Name),
						kv("optionValue", rawValue),
					)
				}
			}
			source = appOptionResultSourceAssign
			found = true
		}
	}

	if !found {
		if declare.DefaultParsedValue != nil {
			rawValue = declare.DefaultRawValue
			parsedValue = declare.DefaultParsedValue
			source = appOptionResultSourceDefault
			found = true
		}
	}

	if parsedValue == nil && !declare.Optional {
		return nil, errN("find option result error",
			reason("option value empty"),
			kv("projectName", manifest.projectName),
			kv("projectPath", manifest.projectPath),
			kv("optionName", declare.Name),
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

	if err = o.addResult(manifest.projectName, declare.Name, result); err != nil {
		return nil, errW(err, "find option result error",
			reason("add option result error"),
			kv("projectName", manifest.projectName),
			kv("projectPath", manifest.projectPath),
			kv("optionName", declare.Name),
		)
	}

	return result, nil
}

// endregion

// region generic items

type appOptionGenericItems map[string]any

const (
	GenericOptionNameOs       = "_os"
	GenericOptionNameArch     = "_arch"
	GenericOptionNameShell    = "_shell"
	GenericOptionNameHostname = "_hostname"
	GenericOptionNameUsername = "_username"
)

func newAppOptionGenericItems(systemInfo *SystemInfo, items map[string]string) appOptionGenericItems {
	os := ""
	if os = items[GenericOptionNameOs]; os == "" {
		os = systemInfo.Os
	}
	arch := ""
	if arch = items[GenericOptionNameArch]; arch == "" {
		arch = systemInfo.Arch
	}
	shell := ""
	if shell = items[GenericOptionNameShell]; shell == "" {
		if os == "windows" {
			shell = "cmd"
		} else {
			shell = "sh"
		}
	}
	hostname := ""
	if hostname = items[GenericOptionNameHostname]; hostname == "" {
		hostname = systemInfo.Os
	}
	username := ""
	if username = items[GenericOptionNameUsername]; username == "" {
		username = systemInfo.Username
	}
	return appOptionGenericItems{
		GenericOptionNameOs:       os,
		GenericOptionNameArch:     arch,
		GenericOptionNameShell:    shell,
		GenericOptionNameHostname: hostname,
		GenericOptionNameUsername: username,
	}
}

func (s appOptionGenericItems) copy() map[string]any {
	result := map[string]any{}
	maps.Copy(result, s)
	return result
}

func (s appOptionGenericItems) merge(items map[string]any) map[string]any {
	result := map[string]any{}
	maps.Copy(result, s)
	if items != nil {
		maps.Copy(result, items)
	}
	return result
}

func (s appOptionGenericItems) getOs() string {
	return s[GenericOptionNameOs].(string)
}

func (s appOptionGenericItems) getShell() string {
	return s[GenericOptionNameShell].(string)
}

// endregion

// region specify items

type appOptionSpecifyItems map[string]map[string]string

func newAppOptionSpecifyItems(projectName string, items map[string]string) appOptionSpecifyItems {
	specifyItems := map[string]string{}
	for k, v := range items {
		if strings.HasPrefix(k, "_") {
			// generic option
			continue
		}
		if strings.Contains(k, ".") {
			// Because of the Option Assign feature
			// You should not specify option values for items other than the main project
			continue
		}
		specifyItems[k] = v
	}
	return map[string]map[string]string{
		projectName: specifyItems,
	}
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
