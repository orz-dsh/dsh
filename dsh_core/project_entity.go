package dsh_core

import (
	"dsh/dsh_utils"
	"github.com/expr-lang/expr/vm"
	"slices"
)

// region option declare

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
	mapping *vm.Program
}

type projectOptionAssignEntitySet []*projectOptionAssignEntity

type projectOptionVerifyEntity struct {
	Expr string
	expr *vm.Program
}

type projectOptionVerifyEntitySet []*projectOptionVerifyEntity

type projectOptionValueType string

const (
	projectOptionValueTypeString  projectOptionValueType = "string"
	projectOptionValueTypeBool    projectOptionValueType = "bool"
	projectOptionValueTypeInteger projectOptionValueType = "integer"
	projectOptionValueTypeDecimal projectOptionValueType = "decimal"
)

func newProjectOptionDeclareEntity(name string, valueType projectOptionValueType, choices []string, optional bool) *projectOptionDeclareEntity {
	return &projectOptionDeclareEntity{
		Name:      name,
		ValueType: valueType,
		Choices:   choices,
		Optional:  optional,
	}
}

func newProjectOptionAssignEntity(project string, option string, mapping string, mappingObj *vm.Program) *projectOptionAssignEntity {
	return &projectOptionAssignEntity{
		Project: project,
		Option:  option,
		Mapping: mapping,
		mapping: mappingObj,
	}
}

func newProjectOptionVerifyEntity(expr string, exprObj *vm.Program) *projectOptionVerifyEntity {
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

// region option specify

type projectOptionSpecifyEntity struct {
	Name  string
	Value string
	Match string
	match *vm.Program
}

type projectOptionSpecifyEntitySet []*projectOptionSpecifyEntity

func newProjectOptionSpecifyEntity(name string, value string, match string, matchObj *vm.Program) *projectOptionSpecifyEntity {
	return &projectOptionSpecifyEntity{
		Name:  name,
		Value: value,
		Match: match,
		match: matchObj,
	}
}

func (s projectOptionSpecifyEntitySet) getItems(evaluator *Evaluator) (map[string]string, error) {
	items := map[string]string{}
	for i := 0; i < len(s); i++ {
		entity := s[i]
		if _, exist := items[entity.Name]; exist {
			continue
		}
		matched, err := evaluator.EvalBoolExpr(entity.match)
		if err != nil {
			return nil, errW(err, "get project option specify items error",
				reason("eval expr error"),
				kv("entity", entity),
			)
		}
		if matched {
			items[entity.Name] = entity.Value
		}
	}
	return items, nil
}

// endregion

// region source

type projectSourceEntity struct {
	Dir   string
	Files []string
	Match string
	match *vm.Program
}

type projectSourceEntitySet []*projectSourceEntity

func newProjectSourceEntity(dir string, files []string, match string, matchObj *vm.Program) *projectSourceEntity {
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
	match *vm.Program
}

type projectImportEntitySet []*projectImportEntity

func newProjectImportEntity(link string, match string, linkObj *projectLink, matchObj *vm.Program) *projectImportEntity {
	return &projectImportEntity{
		Link:  link,
		Match: match,
		link:  linkObj,
		match: matchObj,
	}
}

// endregion
