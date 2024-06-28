package core

import (
	"github.com/orz-dsh/dsh/utils"
	"path/filepath"
)

// region profileSetting

type profileSetting struct {
	Option   *profileOptionSetting
	Project  *profileProjectSetting
	Executor *executorSetting
	Registry *registrySetting
	Redirect *redirectSetting
}

func newProfileSetting(option *profileOptionSetting, project *profileProjectSetting, executor *executorSetting, registry *registrySetting, redirect *redirectSetting) *profileSetting {
	if option == nil {
		option = newProfileOptionSetting(nil)
	}
	if project == nil {
		project = newProfileProjectSetting(nil)
	}
	if executor == nil {
		executor = newExecutorSetting(nil)
	}
	if registry == nil {
		registry = newRegistrySetting(nil)
	}
	if redirect == nil {
		redirect = newRedirectSetting(nil)
	}
	return &profileSetting{
		Option:   option,
		Project:  project,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func loadProfileSetting(file string) (setting *profileSetting, error error) {
	model := &profileSettingModel{}
	metadata, err := utils.DeserializeFromFile(file, "", model)
	if err != nil {
		return nil, errW(err, "load profile setting error",
			reason("deserialize error"),
			kv("file", file),
		)
	}
	if setting, err = model.convert(newModelHelper("profile setting", metadata.File)); err != nil {
		return nil, err
	}
	return setting, nil
}

func loadProfileSettingModel(model *profileSettingModel) (setting *profileSetting, err error) {
	if setting, err = model.convert(newModelHelper("profile setting", "")); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region profileOptionSetting

type profileOptionSetting struct {
	Items []*profileOptionItemSetting
}

func newProfileOptionSetting(items []*profileOptionItemSetting) *profileOptionSetting {
	return &profileOptionSetting{
		Items: items,
	}
}

func (s *profileOptionSetting) merge(setting *profileOptionSetting) {
	s.Items = append(s.Items, setting.Items...)
}

func (s *profileOptionSetting) getItems(evaluator *Evaluator) (map[string]string, error) {
	items := map[string]string{}
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if _, exist := items[item.Name]; exist {
			continue
		}
		matched, err := evaluator.EvalBoolExpr(item.match)
		if err != nil {
			return nil, errW(err, "get profile option specify items error",
				reason("eval expr error"),
				kv("item", item),
			)
		}
		if matched {
			items[item.Name] = item.Value
		}
	}
	return items, nil
}

func (s *profileOptionSetting) inspect() *ProfileOptionSettingInspection {
	var items []*ProfileOptionItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newProfileOptionSettingInspection(items)
}

// endregion

// region profileOptionItemSetting

type profileOptionItemSetting struct {
	Name  string
	Value string
	Match string
	match *EvalExpr
}

func newProfileOptionItemSetting(name, value, match string, matchObj *EvalExpr) *profileOptionItemSetting {
	return &profileOptionItemSetting{
		Name:  name,
		Value: value,
		Match: match,
		match: matchObj,
	}
}

func (s *profileOptionItemSetting) inspect() *ProfileOptionItemSettingInspection {
	return newProfileOptionItemSettingInspection(s.Name, s.Value, s.Match)
}

// endregion

// region profileProjectSetting

type profileProjectSetting struct {
	Items []*profileProjectItemSetting
}

func newProfileProjectSetting(items []*profileProjectItemSetting) *profileProjectSetting {
	return &profileProjectSetting{
		Items: items,
	}
}

func (s *profileProjectSetting) merge(setting *profileProjectSetting) {
	s.Items = append(s.Items, setting.Items...)
}

func (s *profileProjectSetting) getProjectSettings(evaluator *Evaluator) ([]*projectSetting, error) {
	var result []*projectSetting
	for i := len(s.Items) - 1; i >= 0; i-- {
		item := s.Items[i]
		matched, err := evaluator.EvalBoolExpr(item.match)
		if err != nil {
			return nil, errW(err, "get profile project settings error",
				reason("eval expr error"),
				kv("item", item),
			)
		}
		if !matched {
			continue
		}

		rawPath, err := evaluator.EvalStringTemplate(item.Dir)
		if err != nil {
			return nil, errW(err, "get profile project settings error",
				reason("eval template error"),
				kv("item", item),
			)
		}
		path, err := filepath.Abs(rawPath)
		if err != nil {
			return nil, errW(err, "get profile project settings error",
				reason("get abs-path error"),
				kv("item", item),
				kv("rawPath", rawPath),
			)
		}

		result = append(result, newProjectSetting(item.Name, path, nil, nil, item.Dependency, item.Resource))
	}
	return result, nil
}

func (s *profileProjectSetting) inspect() *ProfileProjectSettingInspection {
	var items []*ProfileProjectItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newProfileProjectSettingInspection(items)
}

// endregion

// region profileProjectItemSetting

type profileProjectItemSetting struct {
	Name       string
	Dir        string
	Match      string
	Dependency *projectDependencySetting
	Resource   *projectResourceSetting
	match      *EvalExpr
}

func newProfileProjectItemSetting(name, dir, match string, dependency *projectDependencySetting, resource *projectResourceSetting, matchObj *EvalExpr) *profileProjectItemSetting {
	if dependency == nil {
		dependency = newProjectDependencySetting(nil)
	}
	if resource == nil {
		resource = newProjectResourceSetting(nil)
	}
	return &profileProjectItemSetting{
		Name:       name,
		Dir:        dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
		match:      matchObj,
	}
}

func (s *profileProjectItemSetting) inspect() *ProfileProjectItemSettingInspection {
	return newProfileProjectItemSettingInspection(s.Name, s.Dir, s.Match, s.Dependency.inspect(), s.Resource.inspect())
}

// endregion
