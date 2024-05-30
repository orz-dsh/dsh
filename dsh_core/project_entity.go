package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
	"slices"
)

// region project

type projectEntity struct {
	Name           string
	Path           string
	OptionDeclares projectOptionDeclareEntitySet
	OptionVerifies projectOptionVerifyEntitySet
	ScriptSources  projectSourceEntitySet
	ScriptImports  projectImportEntitySet
	ConfigSources  projectSourceEntitySet
	ConfigImports  projectImportEntitySet
}

type projectEntitySet []*projectEntity

var projectNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*$")

func newProjectEntity(name string, path string, optionDeclares projectOptionDeclareEntitySet, optionVerifies projectOptionVerifyEntitySet, scriptSources projectSourceEntitySet, scriptImports projectImportEntitySet, configSources projectSourceEntitySet, configImports projectImportEntitySet) *projectEntity {
	return &projectEntity{
		Name:           name,
		Path:           path,
		OptionDeclares: optionDeclares,
		OptionVerifies: optionVerifies,
		ScriptSources:  scriptSources,
		ScriptImports:  scriptImports,
		ConfigSources:  configSources,
		ConfigImports:  configImports,
	}
}

// endregion

// region option

type projectOptionDeclareEntity struct {
	Name               string
	ValueType          projectOptionValueType
	Choices            []string
	Optional           bool
	DefaultRawValue    string
	DefaultParsedValue any
	Assigns            projectOptionAssignEntitySet
}

type projectOptionDeclareEntitySet []*projectOptionDeclareEntity

type projectOptionAssignEntity struct {
	Project string
	Option  string
	Mapping string
	mapping *EvalExpr
}

type projectOptionAssignEntitySet []*projectOptionAssignEntity

type projectOptionVerifyEntity struct {
	Expr string
	expr *EvalExpr
}

type projectOptionVerifyEntitySet []*projectOptionVerifyEntity

type projectOptionValueType string

const (
	projectOptionValueTypeString  projectOptionValueType = "string"
	projectOptionValueTypeBool    projectOptionValueType = "bool"
	projectOptionValueTypeInteger projectOptionValueType = "integer"
	projectOptionValueTypeDecimal projectOptionValueType = "decimal"
)

var projectOptionNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*$")

func newProjectOptionDeclareEntity(name string, valueType projectOptionValueType, choices []string, optional bool) *projectOptionDeclareEntity {
	return &projectOptionDeclareEntity{
		Name:      name,
		ValueType: valueType,
		Choices:   choices,
		Optional:  optional,
	}
}

func newProjectOptionAssignEntity(project string, option string, mapping string, mappingObj *EvalExpr) *projectOptionAssignEntity {
	return &projectOptionAssignEntity{
		Project: project,
		Option:  option,
		Mapping: mapping,
		mapping: mappingObj,
	}
}

func newProjectOptionVerifyEntity(expr string, exprObj *EvalExpr) *projectOptionVerifyEntity {
	return &projectOptionVerifyEntity{
		Expr: expr,
		expr: exprObj,
	}
}

func (e *projectOptionDeclareEntity) setDefaultValue(defaultValue *string) error {
	if defaultValue != nil {
		defaultRawValue := *defaultValue
		defaultParsedValue, err := e.parseValue(defaultRawValue)
		if err != nil {
			return err
		}
		e.DefaultRawValue = defaultRawValue
		e.DefaultParsedValue = defaultParsedValue
	}
	return nil
}

func (e *projectOptionDeclareEntity) addAssign(assign *projectOptionAssignEntity) {
	e.Assigns = append(e.Assigns, assign)
}

func (e *projectOptionDeclareEntity) parseValue(rawValue string) (any, error) {
	if len(e.Choices) > 0 && !slices.Contains(e.Choices, rawValue) {
		return nil, errN("option parse value error",
			reason("not in choices"),
			kv("name", e.Name),
			kv("value", rawValue),
			kv("choices", e.Choices),
		)
	}
	var parsedValue any = nil
	switch e.ValueType {
	case projectOptionValueTypeString:
		parsedValue = rawValue
	case projectOptionValueTypeBool:
		parsedValue = rawValue == "true"
	case projectOptionValueTypeInteger:
		integer, err := dsh_utils.ParseInteger(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse integer error"),
				kv("name", e.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = integer
	case projectOptionValueTypeDecimal:
		decimal, err := dsh_utils.ParseDecimal(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse decimal error"),
				kv("name", e.Name),
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

// region source

type projectSourceEntity struct {
	Dir   string
	Files []string
	Match string
	match *EvalExpr
}

type projectSourceEntitySet []*projectSourceEntity

func newProjectSourceEntity(dir string, files []string, match string, matchObj *EvalExpr) *projectSourceEntity {
	return &projectSourceEntity{
		Dir:   dir,
		Files: files,
		Match: match,
		match: matchObj,
	}
}

// endregion

// region import

type projectImportEntity struct {
	Link  string
	Match string
	link  *projectLink
	match *EvalExpr
}

type projectImportEntitySet []*projectImportEntity

func newProjectImportEntity(link string, match string, linkObj *projectLink, matchObj *EvalExpr) *projectImportEntity {
	return &projectImportEntity{
		Link:  link,
		Match: match,
		link:  linkObj,
		match: matchObj,
	}
}

// endregion
