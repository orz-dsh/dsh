package internal

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"maps"
	"strings"
)

// region base

type ApplicationOptionResultSource string

const (
	ApplicationOptionResultSourceUnset   ApplicationOptionResultSource = "unset"
	ApplicationOptionResultSourceExport  ApplicationOptionResultSource = "export"
	ApplicationOptionResultSourceAssign  ApplicationOptionResultSource = "assign"
	ApplicationOptionResultSourceCompute ApplicationOptionResultSource = "compute"
	ApplicationOptionResultSourceDefault ApplicationOptionResultSource = "default"
)

type ApplicationOptionExportSource string

const (
	ApplicationOptionExportSourceAssign  ApplicationOptionExportSource = "assign"
	ApplicationOptionExportSourceCompute ApplicationOptionExportSource = "compute"
	ApplicationOptionExportSourceDefault ApplicationOptionExportSource = "default"
)

// endregion

// region ApplicationOption

type ApplicationOption struct {
	Common *ApplicationOptionCommon
	Export *ApplicationOptionExport
	Assign *ApplicationOptionAssign
	Result *ApplicationOptionResult
}

func NewApplicationOption(projectName string, system *System, evaluator *Evaluator, assigns map[string]string) *ApplicationOption {
	return &ApplicationOption{
		Common: NewApplicationOptionCommon(system, assigns),
		Export: NewApplicationOptionExport(assigns),
		Assign: NewApplicationOptionAssign(projectName, assigns),
		Result: NewApplicationOptionResult(evaluator),
	}
}

func (o *ApplicationOption) findResult(projectName string, setting *ProjectOptionItemSetting) (result *ApplicationOptionResultItem, err error) {
	exportValue, err := o.Export.GetValue(projectName, setting)
	if err != nil {
		return nil, err
	}

	assignValue, err := o.Assign.GetValue(projectName, setting)
	if err != nil {
		return nil, err
	}

	computeValue, err := o.Result.GetComputeValue(projectName, setting)
	if err != nil {
		return nil, err
	}

	if exportValue != nil {
		if assignValue != nil && !exportValue.DeepEqual(assignValue) {
			return nil, ErrN("find option result error",
				Reason("export value conflict"),
				KV("projectName", projectName),
				KV("optionName", setting.Name),
				KV("exportValue", exportValue.Value),
				KV("assignValue", assignValue.Value),
			)
		}
		if computeValue != nil && !exportValue.DeepEqual(computeValue) {
			return nil, ErrN("find option result error",
				Reason("export value conflict"),
				KV("projectName", projectName),
				KV("optionName", setting.Name),
				KV("exportValue", exportValue.Value),
				KV("computeValue", computeValue.Value),
			)
		}
	}

	if assignValue != nil {
		if computeValue != nil && !assignValue.DeepEqual(computeValue) {
			return nil, ErrN("find option result error",
				Reason("assign value conflict"),
				KV("projectName", projectName),
				KV("optionName", setting.Name),
				KV("assignValue", assignValue.Value),
				KV("computeValue", computeValue.Value),
			)
		}
	}

	source := ApplicationOptionResultSourceUnset
	var value any
	if exportValue != nil {
		value = exportValue.Value
		source = ApplicationOptionResultSourceExport
	} else if assignValue != nil {
		value = assignValue.Value
		source = ApplicationOptionResultSourceAssign
	} else if computeValue != nil {
		value = computeValue.Value
		source = ApplicationOptionResultSourceCompute
	} else if setting.Default != nil {
		value = setting.Default
		source = ApplicationOptionResultSourceDefault
	}

	if !setting.Optional && value == nil {
		return nil, ErrN("find option result error",
			Reason("option value empty"),
			KV("projectName", projectName),
			KV("optionName", setting.Name),
			KV("source", source),
		)
	}

	result = NewApplicationOptionResultItem(value, source)
	if err = o.Result.AddItem(projectName, setting.Name, result); err != nil {
		return nil, ErrW(err, "find option result error",
			Reason("add option result error"),
			KV("projectName", projectName),
			KV("optionName", setting.Name),
		)
	}

	if source != ApplicationOptionResultSourceExport {
		o.Export.Items[setting.Export] = NewApplicationOptionExportItem(value, setting.Type, ApplicationOptionExportSource(source))
	}

	return result, nil
}

func (o *ApplicationOption) Inspect() *ApplicationOptionInspection {
	return NewApplicationOptionInspection(o.Common.Inspect(), o.Export.Inspect(), o.Assign.Inspect(), o.Result.Inspect())
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

func NewApplicationOptionCommon(system *System, arguments map[string]string) *ApplicationOptionCommon {
	os := ""
	if os = arguments[OptionNameCommonOs]; os == "" {
		os = system.Os
	}
	arch := ""
	if arch = arguments[OptionNameCommonArch]; arch == "" {
		arch = system.Arch
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
		hostname = system.Os
	}
	username := ""
	if username = arguments[OptionNameCommonUsername]; username == "" {
		username = system.Username
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

// region ApplicationOptionExport

type ApplicationOptionExport struct {
	Items   map[string]*ApplicationOptionExportItem
	Assigns map[string]string
}

func NewApplicationOptionExport(assigns map[string]string) *ApplicationOptionExport {
	a := map[string]string{}
	for k, v := range assigns {
		if strings.Contains(k, ".") {
			a[k] = v
		}
	}
	return &ApplicationOptionExport{
		Items:   map[string]*ApplicationOptionExportItem{},
		Assigns: a,
	}
}

func (e *ApplicationOptionExport) GetValue(projectName string, setting *ProjectOptionItemSetting) (*Value, error) {
	var value *Value
	if export, exist := e.Items[setting.Export]; exist {
		if export.Type != setting.Type {
			return nil, ErrN("find option result error",
				Reason("export value type conflict"),
				KV("projectName", projectName),
				KV("optionName", setting.Name),
				KV("optionType", setting.Type),
				KV("exportValue", export.Value),
				KV("exportType", export.Type),
			)
		}
		value = WrapValue(export.Value)
	}
	if value == nil {
		if assign, exist := e.Assigns[setting.Export]; exist {
			assignValue, err := setting.ParseValue(assign)
			if err != nil {
				return nil, ErrW(err, "find option result error",
					Reason("parse argument value error"),
					KV("optionName", setting.Name),
					KV("optionValue", assign),
				)
			}
			e.Items[setting.Export] = NewApplicationOptionExportItem(assignValue, setting.Type, ApplicationOptionExportSourceAssign)
			value = WrapValue(assignValue)
		}
	}
	return value, nil
}

func (e *ApplicationOptionExport) Inspect() *ApplicationOptionExportInspection {
	items := map[string]*ApplicationOptionExportItemInspection{}
	for k, v := range e.Items {
		items[k] = v.Inspect()
	}
	return NewApplicationOptionExportInspection(items)
}

// endregion

// region ApplicationOptionExportItem

type ApplicationOptionExportItem struct {
	Value  any
	Type   CastType
	Source ApplicationOptionExportSource
}

func NewApplicationOptionExportItem(value any, typ CastType, source ApplicationOptionExportSource) *ApplicationOptionExportItem {
	return &ApplicationOptionExportItem{
		Value:  value,
		Type:   typ,
		Source: source,
	}
}

func (i *ApplicationOptionExportItem) Inspect() *ApplicationOptionExportItemInspection {
	return NewApplicationOptionExportItemInspection(i.Value, string(i.Type), string(i.Source))
}

// endregion

// region ApplicationOptionAssign

type ApplicationOptionAssign struct {
	Items map[string]map[string]string
}

func NewApplicationOptionAssign(projectName string, assigns map[string]string) *ApplicationOptionAssign {
	items := map[string]string{}
	for k, v := range assigns {
		if strings.HasPrefix(k, "_") {
			// common option
			continue
		}
		if strings.Contains(k, ".") {
			// export option
			continue
		}
		items[k] = v
	}
	return &ApplicationOptionAssign{
		Items: map[string]map[string]string{
			projectName: items,
		},
	}
}

func (a *ApplicationOptionAssign) GetValue(projectName string, setting *ProjectOptionItemSetting) (*Value, error) {
	var value *Value
	if assigns, exist := a.Items[projectName]; exist {
		if assign, exist := assigns[setting.Name]; exist {
			assignValue, err := setting.ParseValue(assign)
			if err != nil {
				return nil, ErrW(err, "find option result error",
					Reason("parse argument value error"),
					KV("projectName", projectName),
					KV("optionName", setting.Name),
					KV("optionValue", assignValue),
				)
			}
			value = WrapValue(assignValue)
		}
	}
	return value, nil
}

func (a *ApplicationOptionAssign) Inspect() *ApplicationOptionAssignInspection {
	return NewApplicationOptionAssignInspection(a.Items)
}

// endregion

// region ApplicationOptionResult

type ApplicationOptionResult struct {
	Evaluator *Evaluator
	Items     map[string]map[string]*ApplicationOptionResultItem
}

func NewApplicationOptionResult(evaluator *Evaluator) *ApplicationOptionResult {
	return &ApplicationOptionResult{
		Evaluator: evaluator,
		Items:     map[string]map[string]*ApplicationOptionResultItem{},
	}
}

func (r *ApplicationOptionResult) NewEvaluator(projectName string) *Evaluator {
	data := map[string]any{}
	for k, v := range r.Items[projectName] {
		data[k] = v.Value
	}
	return r.Evaluator.SetRootData("option", data)
}

func (r *ApplicationOptionResult) GetComputeValue(projectName string, setting *ProjectOptionItemSetting) (*Value, error) {
	if setting.Compute == "" {
		return nil, nil
	}
	value, err := setting.ComputeValue(r.NewEvaluator(projectName))
	if err != nil {
		return nil, ErrW(err, "get compute value error",
			KV("projectName", projectName),
			KV("optionName", setting.Name),
		)
	}
	return WrapValue(value), nil
}

func (r *ApplicationOptionResult) AddItem(projectName string, optionName string, result *ApplicationOptionResultItem) error {
	projectItems := r.Items[projectName]
	if projectItems == nil {
		projectItems = map[string]*ApplicationOptionResultItem{}
		r.Items[projectName] = projectItems
	}
	if existResult, exist := projectItems[optionName]; exist {
		return ErrN("add option result error",
			Reason("option result exists"),
			KV("projectName", projectName),
			KV("optionName", optionName),
			KV("result", result),
			KV("existResult", existResult),
		)
	}
	projectItems[optionName] = result
	return nil
}

func (r *ApplicationOptionResult) Inspect() *ApplicationOptionResultInspection {
	items := map[string]map[string]*ApplicationOptionResultItemInspection{}
	for k, v := range r.Items {
		projectItems := map[string]*ApplicationOptionResultItemInspection{}
		for k2, v2 := range v {
			projectItems[k2] = v2.Inspect()
		}
		items[k] = projectItems
	}
	return NewApplicationOptionResultInspection(items)
}

// endregion

// region ApplicationOptionResultItem

type ApplicationOptionResultItem struct {
	Value  any
	Source ApplicationOptionResultSource
}

func NewApplicationOptionResultItem(value any, source ApplicationOptionResultSource) *ApplicationOptionResultItem {
	return &ApplicationOptionResultItem{
		Value:  value,
		Source: source,
	}
}

func (i *ApplicationOptionResultItem) Inspect() *ApplicationOptionResultItemInspection {
	return NewApplicationOptionResultItemInspection(i.Value, string(i.Source))
}

// endregion
