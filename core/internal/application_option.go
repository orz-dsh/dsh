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
	Assign *ApplicationOptionAssign
	Common *ApplicationOptionCommon
	Export *ApplicationOptionExport
	Result *ApplicationOptionResult
}

func NewApplicationOption(projectName string, system *System, evaluator *Evaluator, assigns map[string]string) *ApplicationOption {
	assign := NewApplicationOptionAssign(projectName, assigns)
	common := NewApplicationOptionCommon(system, evaluator, assign)
	return &ApplicationOption{
		Assign: assign,
		Common: common,
		Export: NewApplicationOptionExport(assign),
		Result: NewApplicationOptionResult(common),
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
		export := NewApplicationOptionExportItem(value, setting.Type, ApplicationOptionExportSource(source))
		export.Links = append(export.Links, NewApplicationOptionExportItemLink(projectName, setting.Name))
		o.Export.Items[setting.Export] = export
	}

	return result, nil
}

func (o *ApplicationOption) Inspect() *ApplicationOptionInspection {
	return NewApplicationOptionInspection(o.Assign.Inspect(), o.Common.Inspect(), o.Export.Inspect(), o.Result.Inspect())
}

// endregion

// region ApplicationOptionAssign

type ApplicationOptionAssign struct {
	Common  map[string]string
	Export  map[string]string
	Project map[string]map[string]string
}

func NewApplicationOptionAssign(projectName string, assigns map[string]string) *ApplicationOptionAssign {
	common := map[string]string{}
	export := map[string]string{}
	project := map[string]string{}
	for k, v := range assigns {
		if strings.HasPrefix(k, "_") {
			common[k] = v
		} else if strings.Contains(k, ".") {
			export[k] = v
		} else {
			project[k] = v
		}
	}
	return &ApplicationOptionAssign{
		Common: common,
		Export: export,
		Project: map[string]map[string]string{
			projectName: project,
		},
	}
}

func (a *ApplicationOptionAssign) GetValue(projectName string, setting *ProjectOptionItemSetting) (*Value, error) {
	var value *Value
	if items, exist1 := a.Project[projectName]; exist1 {
		if item, exist2 := items[setting.Name]; exist2 {
			assignValue, err := setting.ParseValue(item)
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

func (a *ApplicationOptionAssign) GetCommonItem(optionName string) (string, bool) {
	if item, exist := a.Common[optionName]; exist {
		return item, true
	}
	return "", false
}

func (a *ApplicationOptionAssign) GetExportItem(exportName string) (string, bool) {
	if item, exist := a.Export[exportName]; exist {
		return item, true
	}
	return "", false
}

func (a *ApplicationOptionAssign) GetProjectItem(projectName, optionName string) (string, bool) {
	if items, exist1 := a.Project[projectName]; exist1 {
		if item, exist2 := items[optionName]; exist2 {
			return item, true
		}
	}
	return "", false
}

func (a *ApplicationOptionAssign) Inspect() *ApplicationOptionAssignInspection {
	return NewApplicationOptionAssignInspection(a.Common, a.Export, a.Project)
}

// endregion

// region ApplicationOptionCommon

type ApplicationOptionCommon struct {
	Evaluator *Evaluator
	Os        string
	Arch      string
	Executor  string
	Hostname  string
	Username  string
}

func NewApplicationOptionCommon(system *System, evaluator *Evaluator, assign *ApplicationOptionAssign) *ApplicationOptionCommon {
	os := ""
	if os, _ = assign.GetCommonItem(OptionNameCommonOs); os == "" {
		os = system.Os
	}
	arch := ""
	if arch, _ = assign.GetCommonItem(OptionNameCommonArch); arch == "" {
		arch = system.Arch
	}
	executor := ""
	if executor, _ = assign.GetCommonItem(OptionNameCommonExecutor); executor == "" {
		if os == "windows" {
			executor = "cmd"
		} else {
			executor = "sh"
		}
	}
	hostname := ""
	if hostname, _ = assign.GetCommonItem(OptionNameCommonHostname); hostname == "" {
		hostname = system.Hostname
	}
	username := ""
	if username, _ = assign.GetCommonItem(OptionNameCommonUsername); username == "" {
		username = system.Username
	}
	return &ApplicationOptionCommon{
		Evaluator: evaluator,
		Os:        os,
		Arch:      arch,
		Executor:  executor,
		Hostname:  hostname,
		Username:  username,
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

func (c *ApplicationOptionCommon) NewEvaluator(items map[string]any) *Evaluator {
	data := c.copy()
	if items != nil {
		maps.Copy(data, items)
	}
	return c.Evaluator.SetRootData("option", data)
}

func (c *ApplicationOptionCommon) Inspect() *ApplicationOptionCommonInspection {
	return NewApplicationOptionCommonInspection(c.Os, c.Arch, c.Executor, c.Hostname, c.Username)
}

// endregion

// region ApplicationOptionExport

type ApplicationOptionExport struct {
	Assign *ApplicationOptionAssign
	Items  map[string]*ApplicationOptionExportItem
}

func NewApplicationOptionExport(assign *ApplicationOptionAssign) *ApplicationOptionExport {
	return &ApplicationOptionExport{
		Assign: assign,
		Items:  map[string]*ApplicationOptionExportItem{},
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
		export.Links = append(export.Links, NewApplicationOptionExportItemLink(projectName, setting.Name))
		value = WrapValue(export.Value)
	}
	if value == nil {
		if assign, exist := e.Assign.GetExportItem(setting.Export); exist {
			assignValue, err := setting.ParseValue(assign)
			if err != nil {
				return nil, ErrW(err, "find option result error",
					Reason("parse argument value error"),
					KV("optionName", setting.Name),
					KV("optionValue", assign),
				)
			}
			export := NewApplicationOptionExportItem(assignValue, setting.Type, ApplicationOptionExportSourceAssign)
			export.Links = append(export.Links, NewApplicationOptionExportItemLink(projectName, setting.Name))
			e.Items[setting.Export] = export
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
	Links  []*ApplicationOptionExportItemLink
}

func NewApplicationOptionExportItem(value any, typ CastType, source ApplicationOptionExportSource) *ApplicationOptionExportItem {
	return &ApplicationOptionExportItem{
		Value:  value,
		Type:   typ,
		Source: source,
	}
}

func (ei *ApplicationOptionExportItem) Inspect() *ApplicationOptionExportItemInspection {
	links := make([]*ApplicationOptionExportItemLinkInspection, 0, len(ei.Links))
	for i := 0; i < len(ei.Links); i++ {
		links = append(links, ei.Links[i].Inspect())
	}
	return NewApplicationOptionExportItemInspection(ei.Value, string(ei.Type), string(ei.Source), links)
}

// endregion

// region ApplicationOptionExportItemLink

type ApplicationOptionExportItemLink struct {
	ProjectName string
	OptionName  string
}

func NewApplicationOptionExportItemLink(projectName, optionName string) *ApplicationOptionExportItemLink {
	return &ApplicationOptionExportItemLink{
		ProjectName: projectName,
		OptionName:  optionName,
	}
}

func (l *ApplicationOptionExportItemLink) Inspect() *ApplicationOptionExportItemLinkInspection {
	return NewApplicationOptionExportItemLinkInspection(l.ProjectName, l.OptionName)
}

// endregion

// region ApplicationOptionResult

type ApplicationOptionResult struct {
	Common *ApplicationOptionCommon
	Items  map[string]map[string]*ApplicationOptionResultItem
}

func NewApplicationOptionResult(common *ApplicationOptionCommon) *ApplicationOptionResult {
	return &ApplicationOptionResult{
		Common: common,
		Items:  map[string]map[string]*ApplicationOptionResultItem{},
	}
}

func (r *ApplicationOptionResult) GetComputeValue(projectName string, setting *ProjectOptionItemSetting) (*Value, error) {
	if setting.Compute == "" {
		return nil, nil
	}
	data := map[string]any{}
	for k, v := range r.Items[projectName] {
		data[k] = v.Value
	}
	value, err := setting.ComputeValue(r.Common.NewEvaluator(data))
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
