package setting

import (
	"encoding/json"
	. "github.com/orz-dsh/dsh/utils"
	"regexp"
	"slices"
	"strings"
)

// region base

var projectOptionNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*[a-z0-9]$")

var projectOptionNameUnsoundDict = map[string]bool{
	"option": true,
	"global": true,
	"local":  true,
}

type ProjectOptionValueType string

const (
	ProjectOptionValueTypeString  ProjectOptionValueType = "string"
	ProjectOptionValueTypeBool    ProjectOptionValueType = "bool"
	ProjectOptionValueTypeInteger ProjectOptionValueType = "integer"
	ProjectOptionValueTypeDecimal ProjectOptionValueType = "decimal"
	ProjectOptionValueTypeObject  ProjectOptionValueType = "object"
	ProjectOptionValueTypeArray   ProjectOptionValueType = "array"
)

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
	Name               string
	ValueType          ProjectOptionValueType
	ValueChoices       []string
	Optional           bool
	DefaultRawValue    string
	DefaultParsedValue any
	Assigns            []*ProjectOptionAssignSetting
}

func NewProjectOptionItemSetting(name string, valueType ProjectOptionValueType, valueChoices []string, optional bool) *ProjectOptionItemSetting {
	return &ProjectOptionItemSetting{
		Name:         name,
		ValueType:    valueType,
		ValueChoices: valueChoices,
		Optional:     optional,
	}
}

func (s *ProjectOptionItemSetting) setDefaultValue(defaultValue *string) error {
	if defaultValue != nil {
		defaultRawValue := *defaultValue
		defaultParsedValue, err := s.ParseValue(defaultRawValue)
		if err != nil {
			return err
		}
		s.DefaultRawValue = defaultRawValue
		s.DefaultParsedValue = defaultParsedValue
	}
	return nil
}

func (s *ProjectOptionItemSetting) addAssign(assign *ProjectOptionAssignSetting) {
	s.Assigns = append(s.Assigns, assign)
}

func (s *ProjectOptionItemSetting) ParseValue(rawValue string) (any, error) {
	if len(s.ValueChoices) > 0 && !slices.Contains(s.ValueChoices, rawValue) {
		return nil, ErrN("parse option value error",
			Reason("not in choices"),
			KV("name", s.Name),
			KV("value", rawValue),
			KV("choices", s.ValueChoices),
		)
	}
	var parsedValue any = nil
	switch s.ValueType {
	case ProjectOptionValueTypeString:
		parsedValue = rawValue
	case ProjectOptionValueTypeBool:
		parsedValue = rawValue == "true"
	case ProjectOptionValueTypeInteger:
		integer, err := ParseInteger(rawValue)
		if err != nil {
			return nil, ErrW(err, "parse option value error",
				Reason("parse integer error"),
				KV("name", s.Name),
				KV("value", rawValue),
			)
		}
		parsedValue = integer
	case ProjectOptionValueTypeDecimal:
		decimal, err := ParseDecimal(rawValue)
		if err != nil {
			return nil, ErrW(err, "parse option value error",
				Reason("parse decimal error"),
				KV("name", s.Name),
				KV("value", rawValue),
			)
		}
		parsedValue = decimal
	case ProjectOptionValueTypeObject:
		var object map[string]any
		if err := json.Unmarshal([]byte(rawValue), &object); err != nil {
			return nil, ErrW(err, "parse option value error",
				Reason("parse object error"),
				KV("name", s.Name),
				KV("value", rawValue),
			)
		}
		parsedValue = object
	case ProjectOptionValueTypeArray:
		var array []any
		if err := json.Unmarshal([]byte(rawValue), &array); err != nil {
			return nil, ErrW(err, "parse option value error",
				Reason("parse array error"),
				KV("name", s.Name),
				KV("value", rawValue),
			)
		}
		parsedValue = array
	default:
		Impossible()
	}
	return parsedValue, nil
}

// endregion

// region ProjectOptionAssignSetting

type ProjectOptionAssignSetting struct {
	Target     string
	Mapping    string
	MappingObj *EvalExpr
}

func NewProjectOptionAssignSetting(target, mapping string, mappingObj *EvalExpr) *ProjectOptionAssignSetting {
	return &ProjectOptionAssignSetting{
		Target:     target,
		Mapping:    mapping,
		MappingObj: mappingObj,
	}
}

// endregion

// region ProjectOptionCheckSetting

type ProjectOptionCheckSetting struct {
	Expr    string
	ExprObj *EvalExpr
}

func NewProjectOptionCheckSetting(expr string, exprObj *EvalExpr) *ProjectOptionCheckSetting {
	return &ProjectOptionCheckSetting{
		Expr:    expr,
		ExprObj: exprObj,
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
	assignTargetsDict := map[string]bool{}
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].Convert(helper.ChildItem("items", i), namesDict, assignTargetsDict)
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
		exprObj, err := CompileExpr(expr)
		if err != nil {
			return nil, helper.ChildItem("checks", i).WrapValueInvalidError(err, expr)
		}
		checks = append(checks, NewProjectOptionCheckSetting(expr, exprObj))
	}

	return NewProjectOptionSetting(items, checks), nil
}

// endregion

// region ProjectOptionItemSettingModel

type ProjectOptionItemSettingModel struct {
	Name     string                             `yaml:"name" toml:"name" json:"name"`
	Type     ProjectOptionValueType             `yaml:"type,omitempty" toml:"type,omitempty" json:"type,omitempty"`
	Choices  []string                           `yaml:"choices,omitempty" toml:"choices,omitempty" json:"choices,omitempty"`
	Default  *string                            `yaml:"default,omitempty" toml:"default,omitempty" json:"default,omitempty"`
	Optional bool                               `yaml:"optional,omitempty" toml:"optional,omitempty" json:"optional,omitempty"`
	Assigns  []*ProjectOptionAssignSettingModel `yaml:"assigns,omitempty" toml:"assigns,omitempty" json:"assigns,omitempty"`
}

func (m *ProjectOptionItemSettingModel) Convert(helper *ModelHelper, namesDict, assignTargetsDict map[string]bool) (*ProjectOptionItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !projectOptionNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}
	if _, exist := namesDict[m.Name]; exist {
		return nil, helper.Child("name").NewError("option name duplicated", KV("name", m.Name))
	}
	if projectOptionNameUnsoundDict[m.Name] {
		helper.Child("name").WarnValueUnsound(m.Name)
	}

	valueType := m.Type
	if valueType == "" {
		valueType = ProjectOptionValueTypeString
	}
	switch valueType {
	case ProjectOptionValueTypeString:
	case ProjectOptionValueTypeBool:
	case ProjectOptionValueTypeInteger:
	case ProjectOptionValueTypeDecimal:
	case ProjectOptionValueTypeObject:
	case ProjectOptionValueTypeArray:
	default:
		return nil, helper.Child("type").NewValueInvalidError(m.Type)
	}

	setting := NewProjectOptionItemSetting(m.Name, valueType, m.Choices, m.Optional)
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

	namesDict[m.Name] = true
	return setting, nil
}

// endregion

// region ProjectOptionAssignSettingModel

type ProjectOptionAssignSettingModel struct {
	Target  string `yaml:"target" toml:"target" json:"target"`
	Mapping string `yaml:"mapping,omitempty" toml:"mapping,omitempty" json:"mapping,omitempty"`
}

func (m *ProjectOptionAssignSettingModel) convert(helper *ModelHelper, targetsDict map[string]bool) (*ProjectOptionAssignSetting, error) {
	if m.Target == "" {
		return nil, helper.Child("target").NewValueEmptyError()
	}

	if strings.Count(m.Target, ".") != 1 {
		return nil, helper.Child("target").NewValueInvalidError(m.Target)
	}

	projectName := helper.GetStringVariable("projectName")
	if projectName != "" && strings.HasPrefix(m.Target, projectName+".") {
		return nil, helper.Child("target").NewError("can not assign to self project option", KV("target", m.Target))
	}

	if _, exists := targetsDict[m.Target]; exists {
		return nil, helper.NewError("option assign target duplicated", KV("target", m.Target))
	}

	mappingObj, err := helper.ConvertEvalExpr("mapping", m.Mapping)
	if err != nil {
		return nil, err
	}

	targetsDict[m.Target] = true
	return NewProjectOptionAssignSetting(m.Target, m.Mapping, mappingObj), nil
}

// endregion
