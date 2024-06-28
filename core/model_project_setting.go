package core

import (
	"github.com/orz-dsh/dsh/utils"
	"regexp"
)

// region base

var projectNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9-]*[a-z0-9]$")

var projectOptionNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*[a-z0-9]$")

type projectOptionValueType string

const (
	projectOptionValueTypeString  projectOptionValueType = "string"
	projectOptionValueTypeBool    projectOptionValueType = "bool"
	projectOptionValueTypeInteger projectOptionValueType = "integer"
	projectOptionValueTypeDecimal projectOptionValueType = "decimal"
	projectOptionValueTypeObject  projectOptionValueType = "object"
	projectOptionValueTypeArray   projectOptionValueType = "array"
)

// endregion

// region projectSettingModel

type projectSettingModel struct {
	Name       string                         `yaml:"name" toml:"name" json:"name"`
	Runtime    *projectRuntimeSettingModel    `yaml:"runtime,omitempty" toml:"runtime,omitempty" json:"runtime,omitempty"`
	Option     *projectOptionSettingModel     `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Dependency *projectDependencySettingModel `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *projectResourceSettingModel   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func (m *projectSettingModel) convert(helper *modelHelper, dir string) (_ *projectSetting, err error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !projectNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}
	helper.AddVariable("projectName", m.Name)

	var runtime *projectRuntimeSetting
	if m.Runtime != nil {
		if runtime, err = m.Runtime.convert(helper.Child("runtime")); err != nil {
			return nil, err
		}
	}

	var option *projectOptionSetting
	if m.Option != nil {
		if option, err = m.Option.convert(helper.Child("option")); err != nil {
			return nil, err
		}
	}

	var dependency *projectDependencySetting
	if m.Dependency != nil {
		if dependency, err = m.Dependency.convert(helper.Child("dependency")); err != nil {
			return nil, err
		}
	}

	var resource *projectResourceSetting
	if m.Resource != nil {
		if resource, err = m.Resource.convert(helper.Child("resource")); err != nil {
			return nil, err
		}
	}

	return newProjectSetting(m.Name, dir, runtime, option, dependency, resource), nil
}

// endregion

// region projectRuntimeSettingModel

type projectRuntimeSettingModel struct {
	MinVersion utils.Version `yaml:"minVersion,omitempty" toml:"minVersion,omitempty" json:"minVersion,omitempty"`
	MaxVersion utils.Version `yaml:"maxVersion,omitempty" toml:"maxVersion,omitempty" json:"maxVersion,omitempty"`
}

func (m *projectRuntimeSettingModel) convert(helper *modelHelper) (*projectRuntimeSetting, error) {
	if err := utils.CheckRuntimeVersion(m.MinVersion, m.MaxVersion); err != nil {
		return nil, helper.WrapError(err, "runtime incompatible",
			kv("minVersion", m.MinVersion),
			kv("maxVersion", m.MaxVersion),
			kv("runtimeVersion", utils.GetRuntimeVersion()),
		)
	}
	return newProjectRuntimeSetting(m.MinVersion, m.MaxVersion), nil
}

// endregion

// region projectOptionSettingModel

type projectOptionSettingModel struct {
	Items  []*projectOptionItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
	Checks []string                         `yaml:"checks,omitempty" toml:"checks,omitempty" json:"checks,omitempty"`
}

func (m *projectOptionSettingModel) convert(helper *modelHelper) (*projectOptionSetting, error) {
	var items []*projectOptionItemSetting
	optionNamesDict := map[string]bool{}
	assignTargetsDict := map[string]bool{}
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].convert(helper.ChildItem("items", i), optionNamesDict, assignTargetsDict)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	var checks []*projectOptionCheckSetting
	for i := 0; i < len(m.Checks); i++ {
		expr := m.Checks[i]
		if expr == "" {
			return nil, helper.ChildItem("checks", i).NewValueEmptyError()
		}
		exprObj, err := utils.CompileExpr(expr)
		if err != nil {
			return nil, helper.ChildItem("checks", i).WrapValueInvalidError(err, expr)
		}
		checks = append(checks, newProjectOptionCheckSetting(expr, exprObj))
	}

	return newProjectOptionSetting(items, checks), nil
}

// endregion

// region projectOptionItemSettingModel

type projectOptionItemSettingModel struct {
	Name     string                                 `yaml:"name" toml:"name" json:"name"`
	Type     projectOptionValueType                 `yaml:"type,omitempty" toml:"type,omitempty" json:"type,omitempty"`
	Choices  []string                               `yaml:"choices,omitempty" toml:"choices,omitempty" json:"choices,omitempty"`
	Default  *string                                `yaml:"default,omitempty" toml:"default,omitempty" json:"default,omitempty"`
	Optional bool                                   `yaml:"optional,omitempty" toml:"optional,omitempty" json:"optional,omitempty"`
	Assigns  []*projectOptionItemAssignSettingModel `yaml:"assigns,omitempty" toml:"assigns,omitempty" json:"assigns,omitempty"`
}

func (m *projectOptionItemSettingModel) convert(helper *modelHelper, itemNamesDict, assignTargetsDict map[string]bool) (*projectOptionItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !projectOptionNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}
	if _, exist := itemNamesDict[m.Name]; exist {
		return nil, helper.Child("name").NewError("option name duplicated", kv("name", m.Name))
	}

	valueType := m.Type
	if valueType == "" {
		valueType = projectOptionValueTypeString
	}
	switch valueType {
	case projectOptionValueTypeString:
	case projectOptionValueTypeBool:
	case projectOptionValueTypeInteger:
	case projectOptionValueTypeDecimal:
	case projectOptionValueTypeObject:
	case projectOptionValueTypeArray:
	default:
		return nil, helper.Child("type").NewValueInvalidError(m.Type)
	}

	setting := newProjectOptionItemSetting(m.Name, valueType, m.Choices, m.Optional)
	if err := setting.setDefaultValue(m.Default); err != nil {
		return nil, helper.Child("default").WrapValueInvalidError(err, *m.Default)
	}

	for i := 0; i < len(m.Assigns); i++ {
		if assignSetting, err := m.Assigns[i].convert(helper.ChildItem("assigns", i), assignTargetsDict); err != nil {
			return nil, err
		} else {
			setting.addAssign(assignSetting)
		}
	}

	itemNamesDict[m.Name] = true
	return setting, nil
}

// endregion

// region projectOptionItemAssignSettingModel

type projectOptionItemAssignSettingModel struct {
	Project string `yaml:"project" toml:"project" json:"project"`
	Option  string `yaml:"option" toml:"option" json:"option"`
	Mapping string `yaml:"mapping,omitempty" toml:"mapping,omitempty" json:"mapping,omitempty"`
}

func (m *projectOptionItemAssignSettingModel) convert(helper *modelHelper, targetsDict map[string]bool) (*projectOptionItemAssignSetting, error) {
	if m.Project == "" {
		return nil, helper.Child("project").NewValueEmptyError()
	}
	if m.Project == helper.GetStringVariable("projectName") {
		return nil, helper.Child("project").NewError("can not assign to self project option")
	}

	if m.Option == "" {
		return nil, helper.Child("option").NewValueEmptyError()
	}

	assignTarget := m.Project + "." + m.Option
	if _, exists := targetsDict[assignTarget]; exists {
		return nil, helper.NewError("option assign target duplicated", kv("target", assignTarget))
	}

	mappingObj, err := helper.ConvertEvalExpr("mapping", m.Mapping)
	if err != nil {
		return nil, err
	}

	targetsDict[assignTarget] = true
	return newProjectOptionItemAssignSetting(m.Project, m.Option, m.Mapping, mappingObj), nil
}

// endregion

// region projectDependencySettingModel

type projectDependencySettingModel struct {
	Items []*projectDependencyItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProjectDependencySettingModel(items []*projectDependencyItemSettingModel) *projectDependencySettingModel {
	return &projectDependencySettingModel{
		Items: items,
	}
}

func (m *projectDependencySettingModel) convert(helper *modelHelper) (*projectDependencySetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newProjectDependencySetting(items), nil
}

// endregion

// region projectDependencyItemSettingModel

type projectDependencyItemSettingModel struct {
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newProjectDependencyItemSettingModel(link, match string) *projectDependencyItemSettingModel {
	return &projectDependencyItemSettingModel{
		Link:  link,
		Match: match,
	}
}

func (m *projectDependencyItemSettingModel) convert(helper *modelHelper) (*projectDependencyItemSetting, error) {
	if m.Link == "" {
		return nil, helper.Child("link").NewValueEmptyError()
	}
	linkObj, err := parseProjectLink(m.Link)
	if err != nil {
		return nil, helper.Child("link").WrapValueInvalidError(err, m.Link)
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newProjectDependencyItemSetting(m.Link, m.Match, linkObj, matchObj), nil
}

// endregion

// region projectResourceSettingModel

type projectResourceSettingModel struct {
	Items []*projectResourceItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProjectResourceSettingModel(items []*projectResourceItemSettingModel) *projectResourceSettingModel {
	return &projectResourceSettingModel{
		Items: items,
	}
}

func (m *projectResourceSettingModel) convert(helper *modelHelper) (*projectResourceSetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newProjectResourceSetting(items), nil
}

// endregion

// region projectResourceItemSettingModel

type projectResourceItemSettingModel struct {
	Dir      string   `yaml:"dir" toml:"dir" json:"dir"`
	Includes []string `yaml:"includes,omitempty" toml:"includes,omitempty" json:"includes,omitempty"`
	Excludes []string `yaml:"excludes,omitempty" toml:"excludes,omitempty" json:"excludes,omitempty"`
	Match    string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newProjectResourceItemSettingModel(dir string, includes, excludes []string, match string) *projectResourceItemSettingModel {
	return &projectResourceItemSettingModel{
		Dir:      dir,
		Includes: includes,
		Excludes: excludes,
		Match:    match,
	}
}

func (m *projectResourceItemSettingModel) convert(helper *modelHelper) (*projectResourceItemSetting, error) {
	if m.Dir == "" {
		return nil, helper.Child("dir").NewValueEmptyError()
	}

	if err := helper.CheckStringItemEmpty("includes", m.Includes); err != nil {
		return nil, err
	}

	if err := helper.CheckStringItemEmpty("excludes", m.Excludes); err != nil {
		return nil, err
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newProjectResourceItemSetting(m.Dir, m.Includes, m.Excludes, m.Match, matchObj), nil
}

// endregion
