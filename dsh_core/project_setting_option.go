package dsh_core

import (
	"dsh/dsh_utils"
	"encoding/json"
	"regexp"
	"slices"
)

// region base

type projectOptionValueType string

const (
	projectOptionValueTypeString  projectOptionValueType = "string"
	projectOptionValueTypeBool    projectOptionValueType = "bool"
	projectOptionValueTypeInteger projectOptionValueType = "integer"
	projectOptionValueTypeDecimal projectOptionValueType = "decimal"
	projectOptionValueTypeObject  projectOptionValueType = "object"
	projectOptionValueTypeArray   projectOptionValueType = "array"
)

var projectOptionNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*[a-z0-9]$")

// endregion

// region projectOptionSetting

type projectOptionSetting struct {
	Items  []*projectOptionItemSetting
	Checks []*projectOptionCheckSetting
}

func newProjectOptionSetting(items []*projectOptionItemSetting, checks []*projectOptionCheckSetting) *projectOptionSetting {
	return &projectOptionSetting{
		Items:  items,
		Checks: checks,
	}
}

// endregion

// region projectOptionItemSetting

type projectOptionItemSetting struct {
	Name               string
	ValueType          projectOptionValueType
	ValueChoices       []string
	Optional           bool
	DefaultRawValue    string
	DefaultParsedValue any
	Assigns            []*projectOptionItemAssignSetting
}

func newProjectOptionItemSetting(name string, valueType projectOptionValueType, valueChoices []string, optional bool) *projectOptionItemSetting {
	return &projectOptionItemSetting{
		Name:         name,
		ValueType:    valueType,
		ValueChoices: valueChoices,
		Optional:     optional,
	}
}

func (s *projectOptionItemSetting) setDefaultValue(defaultValue *string) error {
	if defaultValue != nil {
		defaultRawValue := *defaultValue
		defaultParsedValue, err := s.parseValue(defaultRawValue)
		if err != nil {
			return err
		}
		s.DefaultRawValue = defaultRawValue
		s.DefaultParsedValue = defaultParsedValue
	}
	return nil
}

func (s *projectOptionItemSetting) addAssign(assign *projectOptionItemAssignSetting) {
	s.Assigns = append(s.Assigns, assign)
}

func (s *projectOptionItemSetting) parseValue(rawValue string) (any, error) {
	if len(s.ValueChoices) > 0 && !slices.Contains(s.ValueChoices, rawValue) {
		return nil, errN("option parse value error",
			reason("not in choices"),
			kv("name", s.Name),
			kv("value", rawValue),
			kv("choices", s.ValueChoices),
		)
	}
	var parsedValue any = nil
	switch s.ValueType {
	case projectOptionValueTypeString:
		parsedValue = rawValue
	case projectOptionValueTypeBool:
		parsedValue = rawValue == "true"
	case projectOptionValueTypeInteger:
		integer, err := dsh_utils.ParseInteger(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse integer error"),
				kv("name", s.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = integer
	case projectOptionValueTypeDecimal:
		decimal, err := dsh_utils.ParseDecimal(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse decimal error"),
				kv("name", s.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = decimal
	case projectOptionValueTypeObject:
		var object map[string]any
		if err := json.Unmarshal([]byte(rawValue), &object); err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse object error"),
				kv("name", s.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = object
	case projectOptionValueTypeArray:
		var array []any
		if err := json.Unmarshal([]byte(rawValue), &array); err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse array error"),
				kv("name", s.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = array
	default:
		impossible()
	}
	return parsedValue, nil
}

// endregion

// region projectOptionItemAssignSetting

type projectOptionItemAssignSetting struct {
	Project string
	Option  string
	Mapping string
	mapping *EvalExpr
}

func newProjectOptionItemAssignSetting(project string, option string, mapping string, mappingObj *EvalExpr) *projectOptionItemAssignSetting {
	return &projectOptionItemAssignSetting{
		Project: project,
		Option:  option,
		Mapping: mapping,
		mapping: mappingObj,
	}
}

// endregion

// region projectOptionCheckSetting

type projectOptionCheckSetting struct {
	Expr string
	expr *EvalExpr
}

func newProjectOptionCheckSetting(expr string, exprObj *EvalExpr) *projectOptionCheckSetting {
	return &projectOptionCheckSetting{
		Expr: expr,
		expr: exprObj,
	}
}

// endregion

// region projectOptionSettingModel

type projectOptionSettingModel struct {
	Items  []*projectOptionItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
	Checks []string                         `yaml:"checks,omitempty" toml:"checks,omitempty" json:"checks,omitempty"`
}

func (m *projectOptionSettingModel) convert(ctx *modelConvertContext) (*projectOptionSetting, error) {
	var items []*projectOptionItemSetting
	optionNamesDict := map[string]bool{}
	assignTargetsDict := map[string]bool{}
	for i := 0; i < len(m.Items); i++ {
		if item, err := m.Items[i].convert(ctx.ChildItem("items", i), optionNamesDict, assignTargetsDict); err != nil {
			return nil, err
		} else {
			items = append(items, item)
		}
	}

	var checks []*projectOptionCheckSetting
	for i := 0; i < len(m.Checks); i++ {
		expr := m.Checks[i]
		if expr == "" {
			return nil, ctx.ChildItem("checks", i).NewValueEmptyError()
		}
		exprObj, err := dsh_utils.CompileExpr(expr)
		if err != nil {
			return nil, ctx.ChildItem("checks", i).WrapValueInvalidError(err, expr)
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

func (m *projectOptionItemSettingModel) convert(ctx *modelConvertContext, itemNamesDict, assignTargetsDict map[string]bool) (*projectOptionItemSetting, error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if !projectOptionNameCheckRegex.MatchString(m.Name) {
		return nil, ctx.Child("name").NewValueInvalidError(m.Name)
	}
	if _, exist := itemNamesDict[m.Name]; exist {
		return nil, ctx.Child("name").NewError("option name duplicated", kv("name", m.Name))
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
		return nil, ctx.Child("type").NewValueInvalidError(m.Type)
	}

	setting := newProjectOptionItemSetting(m.Name, valueType, m.Choices, m.Optional)
	if err := setting.setDefaultValue(m.Default); err != nil {
		return nil, ctx.Child("default").WrapValueInvalidError(err, *m.Default)
	}

	for i := 0; i < len(m.Assigns); i++ {
		if assignSetting, err := m.Assigns[i].convert(ctx.ChildItem("assigns", i), assignTargetsDict); err != nil {
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

func (m *projectOptionItemAssignSettingModel) convert(ctx *modelConvertContext, targetsDict map[string]bool) (_ *projectOptionItemAssignSetting, err error) {
	if m.Project == "" {
		return nil, ctx.Child("project").NewValueEmptyError()
	}
	if m.Project == ctx.GetStringVariable("projectName") {
		return nil, ctx.Child("project").NewError("can not assign to self project option")
	}

	if m.Option == "" {
		return nil, ctx.Child("option").NewValueEmptyError()
	}

	assignTarget := m.Project + "." + m.Option
	if _, exists := targetsDict[assignTarget]; exists {
		return nil, ctx.NewError("option assign target duplicated", kv("target", assignTarget))
	}

	var mappingObj *EvalExpr
	if m.Mapping != "" {
		mappingObj, err = dsh_utils.CompileExpr(m.Mapping)
		if err != nil {
			return nil, ctx.Child("mapping").WrapValueInvalidError(err, m.Mapping)
		}
	}

	targetsDict[assignTarget] = true
	return newProjectOptionItemAssignSetting(m.Project, m.Option, m.Mapping, mappingObj), nil
}

// endregion
