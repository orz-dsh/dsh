package dsh_core

import (
	"maps"
	"strings"
)

// region appOption

type appOption struct {
	evaluator    *Evaluator
	GenericItems appOptionGenericItems
	SpecifyItems appOptionSpecifyItems
	Assigns      appOptionAssignSet
	Results      appOptionResultSet
}

func newAppOption(systemInfo *SystemInfo, evaluator *Evaluator, projectName string, specifyItems map[string]string) *appOption {
	return &appOption{
		evaluator:    evaluator,
		GenericItems: newAppOptionGenericItems(systemInfo, specifyItems),
		SpecifyItems: newAppOptionSpecifyItems(projectName, specifyItems),
		Assigns:      appOptionAssignSet{},
		Results:      appOptionResultSet{},
	}
}

func (o *appOption) addAssign(sourceProject string, sourceOption string, assignSetting *projectOptionAssignSetting) error {
	source := sourceProject + "." + sourceOption
	target := assignSetting.Project + "." + assignSetting.Option
	assign := &appOptionAssign{
		Source:      source,
		FinalSource: source,
		Mapping:     assignSetting.Mapping,
		mapping:     assignSetting.mapping,
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

func (o *appOption) findResult(projectName string, declare *projectOptionSetting) (result *appOptionResult, err error) {
	found := false
	var rawValue string
	var parsedValue any = nil
	var source appOptionResultSource = appOptionResultSourceUnset
	var assign *appOptionAssign = nil

	if specifyItems, exist := o.SpecifyItems[projectName]; exist {
		if value, exist := specifyItems[declare.Name]; exist {
			rawValue = value
			parsedValue, err = declare.parseValue(rawValue)
			if err != nil {
				return nil, errW(err, "find option result error",
					reason("parse specify value error"),
					kv("projectName", projectName),
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
		assign, assignValue, err = o.findAssignValue(projectName, declare.Name)
		if err != nil {
			return nil, errW(err, "find option result error",
				reason("get assign value error"),
				kv("projectName", projectName),
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
						kv("projectName", projectName),
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
			kv("projectName", projectName),
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

	if err = o.addResult(projectName, declare.Name, result); err != nil {
		return nil, errW(err, "find option result error",
			reason("add option result error"),
			kv("projectName", projectName),
			kv("optionName", declare.Name),
		)
	}

	return result, nil
}

func (o *appOption) inspect() *AppOptionInspection {
	return newAppOptionInspection(o.GenericItems, o.SpecifyItems, o.Assigns.inspect(), o.Results.inspect())
}

// endregion

// region appOptionGenericItems

type appOptionGenericItems map[string]any

const (
	GenericOptionNameOs       = "_os"
	GenericOptionNameArch     = "_arch"
	GenericOptionNameExecutor = "_executor"
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
	executor := ""
	if executor = items[GenericOptionNameExecutor]; executor == "" {
		if os == "windows" {
			executor = "cmd"
		} else {
			executor = "sh"
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
		GenericOptionNameExecutor: executor,
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

func (s appOptionGenericItems) getExecutor() string {
	return s[GenericOptionNameExecutor].(string)
}

// endregion

// region appOptionSpecifyItems

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

// region appOptionAssign

type appOptionAssign struct {
	Source      string
	FinalSource string
	Mapping     string
	mapping     *EvalExpr
}

type appOptionAssignSet map[string]*appOptionAssign

func (s appOptionAssignSet) inspect() map[string]*AppOptionAssignInspection {
	result := map[string]*AppOptionAssignInspection{}
	for k, v := range s {
		result[k] = newAppOptionAssignInspection(v.Source, v.FinalSource, v.Mapping)
	}
	return result
}

// endregion

// region appOptionResult

type appOptionResult struct {
	RawValue    string
	ParsedValue any
	Source      appOptionResultSource
	Assign      *appOptionAssign
}

type appOptionResultSet map[string]*appOptionResult

type appOptionResultSource string

const (
	appOptionResultSourceUnset   = "unset"
	appOptionResultSourceSpecify = "specify"
	appOptionResultSourceAssign  = "assign"
	appOptionResultSourceDefault = "default"
)

func (s appOptionResultSet) inspect() map[string]*AppOptionResultInspection {
	result := map[string]*AppOptionResultInspection{}
	for k, v := range s {
		var assign *AppOptionAssignInspection
		if v.Assign != nil {
			assign = newAppOptionAssignInspection(v.Assign.Source, v.Assign.FinalSource, v.Assign.Mapping)
		}
		result[k] = newAppOptionResultInspection(v.RawValue, v.ParsedValue, string(v.Source), assign)
	}
	return result
}

// endregion

// region AppOptionInspection

type AppOptionInspection struct {
	GenericItems map[string]any                        `yaml:"genericItems" toml:"genericItems" json:"genericItems"`
	SpecifyItems map[string]map[string]string          `yaml:"specifyItems" toml:"specifyItems" json:"specifyItems"`
	Assigns      map[string]*AppOptionAssignInspection `yaml:"assigns" toml:"assigns" json:"assigns"`
	Results      map[string]*AppOptionResultInspection `yaml:"results" toml:"results" json:"results"`
}

func newAppOptionInspection(genericItems map[string]any, specifyItems map[string]map[string]string, assigns map[string]*AppOptionAssignInspection, results map[string]*AppOptionResultInspection) *AppOptionInspection {
	return &AppOptionInspection{
		GenericItems: genericItems,
		SpecifyItems: specifyItems,
		Assigns:      assigns,
		Results:      results,
	}
}

// endregion

// region AppOptionAssignInspection

type AppOptionAssignInspection struct {
	Source      string `yaml:"source" toml:"source" json:"source"`
	FinalSource string `yaml:"finalSource" toml:"finalSource" json:"finalSource"`
	Mapping     string `yaml:"mapping,omitempty" toml:"mapping,omitempty" json:"mapping,omitempty"`
}

func newAppOptionAssignInspection(source string, finalSource string, mapping string) *AppOptionAssignInspection {
	return &AppOptionAssignInspection{
		Source:      source,
		FinalSource: finalSource,
		Mapping:     mapping,
	}
}

// endregion

// region AppOptionResultInspection

type AppOptionResultInspection struct {
	RawValue    string                     `yaml:"rawValue" toml:"rawValue" json:"rawValue"`
	ParsedValue any                        `yaml:"parsedValue" toml:"parsedValue" json:"parsedValue"`
	Source      string                     `yaml:"source" toml:"source" json:"source"`
	Assign      *AppOptionAssignInspection `yaml:"assign,omitempty" toml:"assign,omitempty" json:"assign,omitempty"`
}

func newAppOptionResultInspection(rawValue string, parsedValue any, source string, assign *AppOptionAssignInspection) *AppOptionResultInspection {
	return &AppOptionResultInspection{
		RawValue:    rawValue,
		ParsedValue: parsedValue,
		Source:      source,
		Assign:      assign,
	}
}

// endregion
