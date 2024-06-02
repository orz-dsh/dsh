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

var projectOptionNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*$")

// endregion

// region option

type projectSchemaOption struct {
	Name               string
	ValueType          projectOptionValueType
	Choices            []string
	Optional           bool
	DefaultRawValue    string
	DefaultParsedValue any
	Assigns            projectSchemaOptionAssignSet
}

type projectSchemaOptionSet []*projectSchemaOption

func newProjectSchemaOption(name string, valueType projectOptionValueType, choices []string, optional bool) *projectSchemaOption {
	return &projectSchemaOption{
		Name:      name,
		ValueType: valueType,
		Choices:   choices,
		Optional:  optional,
	}
}

func (o *projectSchemaOption) setDefaultValue(defaultValue *string) error {
	if defaultValue != nil {
		defaultRawValue := *defaultValue
		defaultParsedValue, err := o.parseValue(defaultRawValue)
		if err != nil {
			return err
		}
		o.DefaultRawValue = defaultRawValue
		o.DefaultParsedValue = defaultParsedValue
	}
	return nil
}

func (o *projectSchemaOption) addAssign(assign *projectSchemaOptionAssign) {
	o.Assigns = append(o.Assigns, assign)
}

func (o *projectSchemaOption) parseValue(rawValue string) (any, error) {
	if len(o.Choices) > 0 && !slices.Contains(o.Choices, rawValue) {
		return nil, errN("option parse value error",
			reason("not in choices"),
			kv("name", o.Name),
			kv("value", rawValue),
			kv("choices", o.Choices),
		)
	}
	var parsedValue any = nil
	switch o.ValueType {
	case projectOptionValueTypeString:
		parsedValue = rawValue
	case projectOptionValueTypeBool:
		parsedValue = rawValue == "true"
	case projectOptionValueTypeInteger:
		integer, err := dsh_utils.ParseInteger(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse integer error"),
				kv("name", o.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = integer
	case projectOptionValueTypeDecimal:
		decimal, err := dsh_utils.ParseDecimal(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse decimal error"),
				kv("name", o.Name),
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

// region option assign

type projectSchemaOptionAssign struct {
	Project string
	Option  string
	Mapping string
	mapping *EvalExpr
}

type projectSchemaOptionAssignSet []*projectSchemaOptionAssign

func newProjectSchemaOptionAssign(project string, option string, mapping string, mappingObj *EvalExpr) *projectSchemaOptionAssign {
	return &projectSchemaOptionAssign{
		Project: project,
		Option:  option,
		Mapping: mapping,
		mapping: mappingObj,
	}
}

// endregion

// region option verify

type projectSchemaOptionVerify struct {
	Expr string
	expr *EvalExpr
}

type projectSchemaOptionVerifySet []*projectSchemaOptionVerify

func newProjectSchemaOptionVerify(expr string, exprObj *EvalExpr) *projectSchemaOptionVerify {
	return &projectSchemaOptionVerify{
		Expr: expr,
		expr: exprObj,
	}
}

// endregion
