package internal

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"maps"
	"strings"
)

// region ApplicationOption

type ApplicationOption struct {
	evaluator *Evaluator
	Common    *ApplicationOptionCommon
	Argument  *ApplicationOptionArgument
	Assign    *ApplicationOptionAssign
	Result    *ApplicationOptionResult
}

func NewApplicationOption(projectName string, systemInfo *SystemInfo, evaluator *Evaluator, arguments map[string]string) *ApplicationOption {
	return &ApplicationOption{
		evaluator: evaluator,
		Common:    NewApplicationOptionCommon(systemInfo, arguments),
		Argument:  NewApplicationOptionArgument(projectName, arguments),
		Assign:    NewApplicationOptionAssign(),
		Result:    NewApplicationOptionResult(),
	}
}

func (o *ApplicationOption) findAssignValue(projectName string, optionName string) (*ApplicationOptionAssignItem, *string, error) {
	target := projectName + "." + optionName
	if assign, exist := o.Assign.Items[target]; exist {
		if result, exist := o.Result.Items[assign.Source]; exist {
			if assign.mapping != nil {
				evaluator := o.evaluator.SetRootData("option", o.Common.merge(map[string]any{
					"value": result.ParsedValue,
				}))
				mappingResult, err := evaluator.EvalStringExpr(assign.mapping)
				if err != nil {
					return assign, nil, ErrW(err, "find option assign value error",
						Reason("mapping value error"),
						KV("target", target),
						KV("assign", assign),
						KV("result", result),
					)
				}
				return assign, mappingResult, nil
			} else {
				return assign, &result.RawValue, nil
			}
		} else {
			return assign, nil, ErrN("find option assign value error",
				Reason("source result not found"),
				KV("target", target),
				KV("assign", assign),
			)
		}
	}
	return nil, nil, nil
}

func (o *ApplicationOption) findResult(projectName string, setting *ProjectOptionItemSetting) (result *ApplicationOptionResultItem, err error) {
	found := false
	var rawValue string
	var parsedValue any = nil
	var source AppOptionResultSource = AppOptionResultSourceUnset
	var assign *ApplicationOptionAssignItem = nil

	if arguments, exist := o.Argument.Items[projectName]; exist {
		if value, exist := arguments[setting.Name]; exist {
			rawValue = value
			parsedValue, err = setting.ParseValue(rawValue)
			if err != nil {
				return nil, ErrW(err, "find option result error",
					Reason("parse argument value error"),
					KV("projectName", projectName),
					KV("optionName", setting.Name),
					KV("optionValue", rawValue),
				)
			}
			source = AppOptionResultSourceArgument
			found = true
		}
	}

	if !found {
		var assignValue *string
		assign, assignValue, err = o.findAssignValue(projectName, setting.Name)
		if err != nil {
			return nil, ErrW(err, "find option result error",
				Reason("get assign value error"),
				KV("projectName", projectName),
				KV("optionName", setting.Name),
			)
		}
		if assign != nil {
			if assignValue != nil {
				rawValue = *assignValue
				parsedValue, err = setting.ParseValue(rawValue)
				if err != nil {
					return nil, ErrW(err, "find option result error",
						Reason("parse assign value error"),
						KV("projectName", projectName),
						KV("optionName", setting.Name),
						KV("optionValue", rawValue),
					)
				}
			}
			source = AppOptionResultSourceAssign
			found = true
		}
	}

	if !found {
		if setting.DefaultParsedValue != nil {
			rawValue = setting.DefaultRawValue
			parsedValue = setting.DefaultParsedValue
			source = AppOptionResultSourceDefault
			found = true
		}
	}

	if parsedValue == nil && !setting.Optional {
		return nil, ErrN("find option result error",
			Reason("option value empty"),
			KV("projectName", projectName),
			KV("optionName", setting.Name),
			KV("source", source),
			KV("assign", assign),
		)
	}

	result = &ApplicationOptionResultItem{
		RawValue:    rawValue,
		ParsedValue: parsedValue,
		Source:      source,
		Assign:      assign,
	}

	if err = o.Result.AddItem(projectName, setting.Name, result); err != nil {
		return nil, ErrW(err, "find option result error",
			Reason("add option result error"),
			KV("projectName", projectName),
			KV("optionName", setting.Name),
		)
	}

	return result, nil
}

func (o *ApplicationOption) Inspect() *ApplicationOptionInspection {
	return NewApplicationOptionInspection(o.Common.Inspect(), o.Argument.Inspect(), o.Assign.Inspect(), o.Result.Inspect())
}

// endregion

// region ApplicationOptionCommon

type ApplicationOptionCommon struct {
	Os       string
	Arch     string
	Executor string
	Hostname string
	Username string
}

func NewApplicationOptionCommon(systemInfo *SystemInfo, arguments map[string]string) *ApplicationOptionCommon {
	os := ""
	if os = arguments[OptionNameCommonOs]; os == "" {
		os = systemInfo.Os
	}
	arch := ""
	if arch = arguments[OptionNameCommonArch]; arch == "" {
		arch = systemInfo.Arch
	}
	executor := ""
	if executor = arguments[OptionNameCommonExecutor]; executor == "" {
		if os == "windows" {
			executor = "cmd"
		} else {
			executor = "sh"
		}
	}
	hostname := ""
	if hostname = arguments[OptionNameCommonHostname]; hostname == "" {
		hostname = systemInfo.Os
	}
	username := ""
	if username = arguments[OptionNameCommonUsername]; username == "" {
		username = systemInfo.Username
	}
	return &ApplicationOptionCommon{
		Os:       os,
		Arch:     arch,
		Executor: executor,
		Hostname: hostname,
		Username: username,
	}
}

func (c *ApplicationOptionCommon) copy() map[string]any {
	return map[string]any{
		OptionNameCommonOs:       c.Os,
		OptionNameCommonArch:     c.Arch,
		OptionNameCommonExecutor: c.Executor,
		OptionNameCommonHostname: c.Hostname,
		OptionNameCommonUsername: c.Username,
	}
}

func (c *ApplicationOptionCommon) merge(items map[string]any) map[string]any {
	result := c.copy()
	if items != nil {
		maps.Copy(result, items)
	}
	return result
}

func (c *ApplicationOptionCommon) Inspect() *ApplicationOptionCommonInspection {
	return NewApplicationOptionCommonInspection(c.Os, c.Arch, c.Executor, c.Hostname, c.Username)
}

// endregion

// region ApplicationOptionArgument

type ApplicationOptionArgument struct {
	Items map[string]map[string]string
}

func NewApplicationOptionArgument(projectName string, arguments map[string]string) *ApplicationOptionArgument {
	items := map[string]string{}
	for k, v := range arguments {
		if strings.HasPrefix(k, "_") {
			// universal option
			continue
		}
		if strings.Contains(k, ".") {
			// Because of the Option Assign feature
			// Should not set option arguments for other than the main project
			continue
		}
		items[k] = v
	}
	return &ApplicationOptionArgument{
		Items: map[string]map[string]string{
			projectName: items,
		},
	}
}

func (a *ApplicationOptionArgument) Inspect() *ApplicationOptionArgumentInspection {
	return NewApplicationOptionArgumentInspection(a.Items)
}

// endregion

// region ApplicationOptionAssign

type ApplicationOptionAssign struct {
	Items map[string]*ApplicationOptionAssignItem
}

func NewApplicationOptionAssign() *ApplicationOptionAssign {
	return &ApplicationOptionAssign{
		Items: map[string]*ApplicationOptionAssignItem{},
	}
}

func (a *ApplicationOptionAssign) AddItem(sourceProject string, sourceOption string, assignSetting *ProjectOptionAssignSetting) error {
	source := sourceProject + "." + sourceOption
	target := assignSetting.Target
	assign := &ApplicationOptionAssignItem{
		Source:      source,
		FinalSource: source,
		Mapping:     assignSetting.Mapping,
		mapping:     assignSetting.MappingObj,
	}
	if sourceAssign, exist := a.Items[source]; exist {
		assign.FinalSource = sourceAssign.FinalSource
	}
	if existAssign, exist := a.Items[target]; exist {
		if existAssign.FinalSource != assign.FinalSource {
			return ErrN("add option assign error",
				Reason("option assign conflict"),
				KV("target", target),
				KV("assign", assign),
				KV("existAssign", existAssign),
			)
		}
	} else {
		a.Items[target] = assign
	}
	return nil
}

func (a *ApplicationOptionAssign) Inspect() *ApplicationOptionAssignInspection {
	items := map[string]*ApplicationOptionAssignItemInspection{}
	for k, v := range a.Items {
		items[k] = v.Inspect()
	}
	return NewApplicationOptionAssignInspection(items)
}

// endregion

// region ApplicationOptionAssignItem

type ApplicationOptionAssignItem struct {
	Source      string
	FinalSource string
	Mapping     string
	mapping     *EvalExpr
}

func (i *ApplicationOptionAssignItem) Inspect() *ApplicationOptionAssignItemInspection {
	return NewApplicationOptionAssignItemInspection(i.Source, i.FinalSource, i.Mapping)
}

// endregion

// region ApplicationOptionResult

type ApplicationOptionResult struct {
	Items map[string]*ApplicationOptionResultItem
}

type AppOptionResultSource string

const (
	AppOptionResultSourceUnset    = "unset"
	AppOptionResultSourceArgument = "argument"
	AppOptionResultSourceAssign   = "assign"
	AppOptionResultSourceDefault  = "default"
)

func NewApplicationOptionResult() *ApplicationOptionResult {
	return &ApplicationOptionResult{
		Items: map[string]*ApplicationOptionResultItem{},
	}
}

func (r *ApplicationOptionResult) AddItem(projectName string, optionName string, result *ApplicationOptionResultItem) error {
	target := projectName + "." + optionName
	if existResult, exist := r.Items[target]; exist {
		return ErrN("add option result error",
			Reason("option result exists"),
			KV("target", target),
			KV("result", result),
			KV("existResult", existResult),
		)
	}
	r.Items[target] = result
	return nil
}

func (r *ApplicationOptionResult) Inspect() *ApplicationOptionResultInspection {
	items := map[string]*ApplicationOptionResultItemInspection{}
	for k, v := range r.Items {
		items[k] = v.Inspect()
	}
	return NewApplicationOptionResultInspection(items)
}

// endregion

// region ApplicationOptionResultItem

type ApplicationOptionResultItem struct {
	RawValue    string
	ParsedValue any
	Source      AppOptionResultSource
	Assign      *ApplicationOptionAssignItem
}

func (i *ApplicationOptionResultItem) Inspect() *ApplicationOptionResultItemInspection {
	var assign *ApplicationOptionAssignItemInspection
	if i.Assign != nil {
		assign = i.Assign.Inspect()
	}
	return NewApplicationOptionResultItemInspection(i.RawValue, i.ParsedValue, string(i.Source), assign)
}

// endregion
