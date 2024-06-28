package core

import (
	"encoding/json"
	"github.com/orz-dsh/dsh/utils"
	"slices"
)

// region projectSetting

type projectSetting struct {
	Name       string
	Dir        string
	Runtime    *projectRuntimeSetting
	Option     *projectOptionSetting
	Dependency *projectDependencySetting
	Resource   *projectResourceSetting
}

func newProjectSetting(name, dir string, runtime *projectRuntimeSetting, option *projectOptionSetting, dependency *projectDependencySetting, resource *projectResourceSetting) *projectSetting {
	if runtime == nil {
		runtime = newProjectRuntimeSetting("", "")
	}
	if option == nil {
		option = newProjectOptionSetting(nil, nil)
	}
	if dependency == nil {
		dependency = newProjectDependencySetting(nil)
	}
	if resource == nil {
		resource = newProjectResourceSetting(nil)
	}
	return &projectSetting{
		Name:       name,
		Dir:        dir,
		Runtime:    runtime,
		Option:     option,
		Dependency: dependency,
		Resource:   resource,
	}
}

func loadProjectSetting(dir string) (setting *projectSetting, err error) {
	model := &projectSettingModel{}
	metadata, err := utils.DeserializeFromDir(dir, []string{"project"}, model, true)
	if err != nil {
		return nil, errW(err, "load project setting error",
			reason("deserialize error"),
			kv("dir", dir),
		)
	}
	if setting, err = model.convert(newModelHelper("project setting", metadata.File), dir); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region projectRuntimeSetting

type projectRuntimeSetting struct {
	MinVersion utils.Version
	MaxVersion utils.Version
	minVersion int32
}

func newProjectRuntimeSetting(minVersion utils.Version, maxVersion utils.Version) *projectRuntimeSetting {
	return &projectRuntimeSetting{
		MinVersion: minVersion,
		MaxVersion: maxVersion,
	}
}

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
		integer, err := utils.ParseInteger(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse integer error"),
				kv("name", s.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = integer
	case projectOptionValueTypeDecimal:
		decimal, err := utils.ParseDecimal(rawValue)
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

// region projectDependencySetting

type projectDependencySetting struct {
	Items []*projectDependencyItemSetting
}

func newProjectDependencySetting(items []*projectDependencyItemSetting) *projectDependencySetting {
	return &projectDependencySetting{
		Items: items,
	}
}

func (s *projectDependencySetting) inspect() *ProjectDependencySettingInspection {
	var items []*ProjectDependencyItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newProjectDependencySettingInspection(items)
}

// endregion

// region projectDependencyItemSetting

type projectDependencyItemSetting struct {
	Link  string
	Match string
	link  *projectLink
	match *EvalExpr
}

func newProjectDependencyItemSetting(link string, match string, linkObj *projectLink, matchObj *EvalExpr) *projectDependencyItemSetting {
	return &projectDependencyItemSetting{
		Link:  link,
		Match: match,
		link:  linkObj,
		match: matchObj,
	}
}

func (s *projectDependencyItemSetting) inspect() *ProjectDependencyItemSettingInspection {
	return newProjectDependencyItemSettingInspection(s.Link, s.Match)
}

// endregion

// region projectResourceSetting

type projectResourceSetting struct {
	Items []*projectResourceItemSetting
}

func newProjectResourceSetting(items []*projectResourceItemSetting) *projectResourceSetting {
	return &projectResourceSetting{
		Items: items,
	}
}

func (s *projectResourceSetting) inspect() *ProjectResourceSettingInspection {
	var items []*ProjectResourceItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newProjectResourceSettingInspection(items)
}

// endregion

// region projectResourceItemSetting

type projectResourceItemSetting struct {
	Dir      string
	Includes []string
	Excludes []string
	Match    string
	match    *EvalExpr
}

func newProjectResourceItemSetting(dir string, includes, excludes []string, match string, matchObj *EvalExpr) *projectResourceItemSetting {
	return &projectResourceItemSetting{
		Dir:      dir,
		Includes: includes,
		Excludes: excludes,
		Match:    match,
		match:    matchObj,
	}
}

func (s *projectResourceItemSetting) inspect() *ProjectResourceItemSettingInspection {
	return newProjectResourceItemSettingInspection(s.Dir, s.Includes, s.Excludes, s.Match)
}

// endregion
