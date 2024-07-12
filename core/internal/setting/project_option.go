package setting

import (
	. "github.com/orz-dsh/dsh/utils"
	"reflect"
	"regexp"
)

// region base

var projectOptionNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*[a-z0-9]$")

var projectOptionExportCheckRegex = regexp.MustCompile("^@?[a-z][a-z0-9_]*[a-z0-9].[a-z][a-z0-9_]*[a-z0-9]$")

var projectOptionNameUnsoundDict = map[string]bool{
	"option": true,
	"global": true,
	"local":  true,
}

type ProjectOptionValueType = CastType

// endregion

// region ProjectOptionSetting

type ProjectOptionSetting struct {
	Items  []*ProjectOptionItemSetting
	Checks []*ProjectOptionCheckSetting
}

func NewProjectOptionSetting(items []*ProjectOptionItemSetting, checks []*ProjectOptionCheckSetting) *ProjectOptionSetting {
	return &ProjectOptionSetting{
		Items:  items,
		Checks: checks,
	}
}

// endregion

// region ProjectOptionItemSetting

type ProjectOptionItemSetting struct {
	Name     string
	Type     ProjectOptionValueType
	Usage    string
	Export   string
	Hidden   bool
	Compute  string
	Default  any
	Choices  []any
	Optional bool
}

func NewProjectOptionItemSetting(name string, typ ProjectOptionValueType, usage, export string, hidden bool, compute string, optional bool) *ProjectOptionItemSetting {
	return &ProjectOptionItemSetting{
		Name:     name,
		Type:     typ,
		Usage:    usage,
		Export:   export,
		Hidden:   hidden,
		Compute:  compute,
		Optional: optional,
	}
}

func (s *ProjectOptionItemSetting) setChoices(choices []string) error {
	if choices != nil {
		result, err := CastSlice(choices, s.Type)
		if err != nil {
			return ErrW(err, "parse option choices error",
				Reason("cast error"),
				KV("name", s.Name),
				KV("choices", choices),
				KV("type", s.Type),
			)
		}
		s.Choices = result
	}
	return nil
}

func (s *ProjectOptionItemSetting) checkChoices(value any) error {
	if len(s.Choices) > 0 {
		for i := 0; i < len(s.Choices); i++ {
			if reflect.DeepEqual(s.Choices[i], value) {
				return nil
			}
		}
		return ErrN("check option choices error",
			Reason("not in choices"),
			KV("name", s.Name),
			KV("value", value),
			KV("choices", s.Choices),
		)
	}
	return nil
}

func (s *ProjectOptionItemSetting) setDefault(value *string) error {
	if value != nil {
		result, err := Cast(*value, s.Type)
		if err != nil {
			return ErrW(err, "parse option value error",
				Reason("cast error"),
				KV("name", s.Name),
				KV("value", value),
				KV("type", s.Type),
			)
		}
		if err = s.checkChoices(result); err != nil {
			return ErrW(err, "option default error", Reason("check choices error"))
		}
		s.Default = result
	}
	return nil
}

func (s *ProjectOptionItemSetting) ParseValue(value string) (any, error) {
	result, err := Cast(value, s.Type)
	if err != nil {
		return nil, ErrW(err, "parse option value error",
			Reason("cast value error"),
			KV("name", s.Name),
			KV("value", value),
			KV("type", s.Type),
		)
	}
	if err = s.checkChoices(result); err != nil {
		return nil, ErrW(err, "parse option value error", Reason("check choices error"))
	}
	return result, nil
}

func (s *ProjectOptionItemSetting) ComputeValue(evaluator *Evaluator) (any, error) {
	result, err := evaluator.EvalExpr(s.Compute, s.Type)
	if err != nil {
		return nil, ErrW(err, "compute option value error",
			Reason("eval expr error"),
			KV("name", s.Name),
			KV("compute", s.Compute),
			KV("type", s.Type),
		)
	}
	if err = s.checkChoices(result); err != nil {
		return nil, ErrW(err, "compute option value error", Reason("check choices error"))
	}
	return result, nil
}

// endregion

// region ProjectOptionCheckSetting

type ProjectOptionCheckSetting struct {
	Expr string
}

func NewProjectOptionCheckSetting(expr string) *ProjectOptionCheckSetting {
	return &ProjectOptionCheckSetting{
		Expr: expr,
	}
}

// endregion

// region ProjectOptionSettingModel

type ProjectOptionSettingModel struct {
	Items  []*ProjectOptionItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
	Checks []string                         `yaml:"checks,omitempty" toml:"checks,omitempty" json:"checks,omitempty"`
}

func (m *ProjectOptionSettingModel) Convert(helper *ModelHelper) (*ProjectOptionSetting, error) {
	var items []*ProjectOptionItemSetting
	namesDict := map[string]bool{}
	exportsDict := map[string]bool{}
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].Convert(helper.ChildItem("items", i), namesDict, exportsDict)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	var checks []*ProjectOptionCheckSetting
	for i := 0; i < len(m.Checks); i++ {
		expr := m.Checks[i]
		if expr == "" {
			return nil, helper.ChildItem("checks", i).NewValueEmptyError()
		}
		checks = append(checks, NewProjectOptionCheckSetting(expr))
	}

	return NewProjectOptionSetting(items, checks), nil
}

// endregion

// region ProjectOptionItemSettingModel

type ProjectOptionItemSettingModel struct {
	Name     string                 `yaml:"name" toml:"name" json:"name"`
	Type     ProjectOptionValueType `yaml:"type,omitempty" toml:"type,omitempty" json:"type,omitempty"`
	Usage    string                 `yaml:"usage,omitempty" toml:"usage,omitempty" json:"usage,omitempty"`
	Export   string                 `yaml:"export,omitempty" toml:"export,omitempty" json:"export,omitempty"`
	Hidden   bool                   `yaml:"hidden,omitempty" toml:"hidden,omitempty" json:"hidden,omitempty"`
	Compute  string                 `yaml:"compute,omitempty" toml:"compute,omitempty" json:"compute,omitempty"`
	Default  *string                `yaml:"default,omitempty" toml:"default,omitempty" json:"default,omitempty"`
	Choices  []string               `yaml:"choices,omitempty" toml:"choices,omitempty" json:"choices,omitempty"`
	Optional bool                   `yaml:"optional,omitempty" toml:"optional,omitempty" json:"optional,omitempty"`
}

func (m *ProjectOptionItemSettingModel) Convert(helper *ModelHelper, namesDict, exportsDict map[string]bool) (*ProjectOptionItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !projectOptionNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}
	if namesDict[m.Name] {
		return nil, helper.Child("name").NewError("option name duplicated", KV("name", m.Name))
	}
	if projectOptionNameUnsoundDict[m.Name] {
		helper.Child("name").WarnValueUnsound(m.Name)
	}

	typ := m.Type
	if typ == "" {
		typ = CastTypeString
	}
	switch typ {
	case CastTypeString:
	case CastTypeBool:
	case CastTypeInteger:
	case CastTypeDecimal:
	case CastTypeObject:
	case CastTypeArray:
	default:
		return nil, helper.Child("type").NewValueInvalidError(typ)
	}

	projectName := helper.GetStringVariable("projectName")
	export := m.Export
	if export == "" {
		export = projectName + "." + m.Name
	}
	if !projectOptionExportCheckRegex.MatchString(export) {
		return nil, helper.Child("export").NewValueInvalidError(export)
	}
	if exportsDict[export] {
		return nil, helper.Child("export").NewError("option export duplicated", KV("export", export))
	}

	setting := NewProjectOptionItemSetting(m.Name, typ, m.Usage, export, m.Hidden, m.Compute, m.Optional)

	if err := setting.setChoices(m.Choices); err != nil {
		return nil, helper.Child("choices").WrapValueInvalidError(err, m.Choices)
	}

	if setting.Compute == "" {
		if err := setting.setDefault(m.Default); err != nil {
			return nil, helper.Child("default").WrapValueInvalidError(err, m.Default)
		}
	} else if m.Default != nil {
		return nil, helper.Child("default").NewError("option compute and default conflict")
	}

	namesDict[m.Name] = true
	exportsDict[export] = true
	return setting, nil
}

// endregion
