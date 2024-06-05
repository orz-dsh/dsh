package dsh_core

import (
	"dsh/dsh_utils"
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
)

var projectOptionNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*[a-z0-9]$")

// endregion

// region projectOptionSetting

type projectOptionSetting struct {
	Name               string
	ValueType          projectOptionValueType
	ValueChoices       []string
	Optional           bool
	DefaultRawValue    string
	DefaultParsedValue any
	AssignSettings     projectOptionAssignSettingSet
}

type projectOptionSettingSet []*projectOptionSetting

func newProjectOptionSetting(name string, valueType projectOptionValueType, valueChoices []string, optional bool) *projectOptionSetting {
	return &projectOptionSetting{
		Name:         name,
		ValueType:    valueType,
		ValueChoices: valueChoices,
		Optional:     optional,
	}
}

func (s *projectOptionSetting) setDefaultValue(defaultValue *string) error {
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

func (s *projectOptionSetting) addAssignSetting(setting *projectOptionAssignSetting) {
	s.AssignSettings = append(s.AssignSettings, setting)
}

func (s *projectOptionSetting) parseValue(rawValue string) (any, error) {
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
	default:
		impossible()
	}
	return parsedValue, nil
}

// endregion

// region projectOptionAssignSetting

type projectOptionAssignSetting struct {
	Project string
	Option  string
	Mapping string
	mapping *EvalExpr
}

type projectOptionAssignSettingSet []*projectOptionAssignSetting

func newProjectOptionAssignSetting(project string, option string, mapping string, mappingObj *EvalExpr) *projectOptionAssignSetting {
	return &projectOptionAssignSetting{
		Project: project,
		Option:  option,
		Mapping: mapping,
		mapping: mappingObj,
	}
}

// endregion

// region projectOptionVerifySetting

type projectOptionVerifySetting struct {
	Expr string
	expr *EvalExpr
}

type projectOptionVerifySettingSet []*projectOptionVerifySetting

func newProjectOptionVerifySetting(expr string, exprObj *EvalExpr) *projectOptionVerifySetting {
	return &projectOptionVerifySetting{
		Expr: expr,
		expr: exprObj,
	}
}

// endregion

// region projectOptionSettingModel

type projectOptionSettingModel struct {
	Items    []*projectOptionItemSettingModel
	Verifies []string
}

func (m *projectOptionSettingModel) convert(ctx *modelConvertContext) (projectOptionSettingSet, projectOptionVerifySettingSet, error) {
	optionSettings := projectOptionSettingSet{}
	optionNamesDict := map[string]bool{}
	assignTargetsDict := map[string]bool{}
	for i := 0; i < len(m.Items); i++ {
		if setting, err := m.Items[i].convert(ctx.ChildItem("items", i), optionNamesDict, assignTargetsDict); err != nil {
			return nil, nil, err
		} else {
			optionSettings = append(optionSettings, setting)
		}
	}

	optionVerifySettings := projectOptionVerifySettingSet{}
	for i := 0; i < len(m.Verifies); i++ {
		expr := m.Verifies[i]
		if expr == "" {
			return nil, nil, ctx.ChildItem("verifies", i).NewValueEmptyError()
		}
		exprObj, err := dsh_utils.CompileExpr(expr)
		if err != nil {
			return nil, nil, ctx.ChildItem("verifies", i).WrapValueInvalidError(err, expr)
		}
		optionVerifySettings = append(optionVerifySettings, newProjectOptionVerifySetting(expr, exprObj))
	}

	return optionSettings, optionVerifySettings, nil
}

// endregion

// region projectOptionItemSettingModel

type projectOptionItemSettingModel struct {
	Name     string
	Type     projectOptionValueType
	Choices  []string
	Default  *string
	Optional bool
	Assigns  []*projectOptionItemAssignSettingModel
}

func (m *projectOptionItemSettingModel) convert(ctx *modelConvertContext, itemNamesDict, assignTargetsDict map[string]bool) (setting *projectOptionSetting, err error) {
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
	default:
		return nil, ctx.Child("type").NewValueInvalidError(m.Type)
	}

	setting = newProjectOptionSetting(m.Name, valueType, m.Choices, m.Optional)
	if err = setting.setDefaultValue(m.Default); err != nil {
		return nil, ctx.Child("default").WrapValueInvalidError(err, *m.Default)
	}

	for i := 0; i < len(m.Assigns); i++ {
		if assignSetting, err := m.Assigns[i].convert(ctx.ChildItem("assigns", i), assignTargetsDict); err != nil {
			return nil, err
		} else {
			setting.addAssignSetting(assignSetting)
		}
	}

	itemNamesDict[m.Name] = true
	return setting, nil
}

// endregion

// region projectOptionItemAssignSettingModel

type projectOptionItemAssignSettingModel struct {
	Project string
	Option  string
	Mapping string
}

func (m *projectOptionItemAssignSettingModel) convert(ctx *modelConvertContext, targetsDict map[string]bool) (setting *projectOptionAssignSetting, err error) {
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
	return newProjectOptionAssignSetting(m.Project, m.Option, m.Mapping, mappingObj), nil
}

// endregion
