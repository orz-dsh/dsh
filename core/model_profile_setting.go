package core

import "regexp"

// region base

var profileOptionNameCheckRegex = regexp.MustCompile("^_?[a-z][a-z0-9_]*[a-z]$")

// endregion

// region profileSettingModel

type profileSettingModel struct {
	Option   *profileOptionSettingModel  `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Project  *profileProjectSettingModel `yaml:"project,omitempty" toml:"project,omitempty" json:"project,omitempty"`
	Executor *executorSettingModel       `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *registrySettingModel       `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *redirectSettingModel       `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
}

func newProfileSettingModel(option *profileOptionSettingModel, project *profileProjectSettingModel, executor *executorSettingModel, registry *registrySettingModel, redirect *redirectSettingModel) *profileSettingModel {
	return &profileSettingModel{
		Option:   option,
		Project:  project,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func (m *profileSettingModel) convert(helper *modelHelper) (_ *profileSetting, err error) {
	var option *profileOptionSetting
	if m.Option != nil {
		if option, err = m.Option.convert(helper.Child("option")); err != nil {
			return nil, err
		}
	}

	var project *profileProjectSetting
	if m.Project != nil {
		if project, err = m.Project.convert(helper.Child("project")); err != nil {
			return nil, err
		}
	}

	var executor *executorSetting
	if m.Executor != nil {
		if executor, err = m.Executor.convert(helper.Child("executor")); err != nil {
			return nil, err
		}
	}

	var registry *registrySetting
	if m.Registry != nil {
		if registry, err = m.Registry.convert(helper.Child("registry")); err != nil {
			return nil, err
		}
	}

	var redirect *redirectSetting
	if m.Redirect != nil {
		if redirect, err = m.Redirect.convert(helper.Child("redirect")); err != nil {
			return nil, err
		}
	}

	return newProfileSetting(option, project, executor, registry, redirect), nil
}

// endregion

// region profileOptionSettingModel

type profileOptionSettingModel struct {
	Items []*profileOptionItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProfileOptionSettingModel(items []*profileOptionItemSettingModel) *profileOptionSettingModel {
	return &profileOptionSettingModel{
		Items: items,
	}
}

func (m *profileOptionSettingModel) convert(helper *modelHelper) (*profileOptionSetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newProfileOptionSetting(items), nil
}

// endregion

// region profileOptionItemSettingModel

type profileOptionItemSettingModel struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func (m *profileOptionItemSettingModel) convert(helper *modelHelper) (*profileOptionItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !profileOptionNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newProfileOptionItemSetting(m.Name, m.Value, m.Match, matchObj), nil
}

// endregion

// region profileProjectSettingModel

type profileProjectSettingModel struct {
	Items []*profileProjectItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProfileProjectSettingModel(items []*profileProjectItemSettingModel) *profileProjectSettingModel {
	return &profileProjectSettingModel{
		Items: items,
	}
}

func (m *profileProjectSettingModel) convert(helper *modelHelper) (*profileProjectSetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newProfileProjectSetting(items), nil
}

// endregion

// region profileProjectItemSettingModel

type profileProjectItemSettingModel struct {
	Name       string                         `yaml:"name" toml:"name" json:"name"`
	Dir        string                         `yaml:"dir" toml:"dir" json:"dir"`
	Match      string                         `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
	Dependency *projectDependencySettingModel `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *projectResourceSettingModel   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func newProfileProjectItemSettingModel(name, dir, match string, dependency *projectDependencySettingModel, resource *projectResourceSettingModel) *profileProjectItemSettingModel {
	return &profileProjectItemSettingModel{
		Name:       name,
		Dir:        dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
	}
}

func (m *profileProjectItemSettingModel) convert(helper *modelHelper) (_ *profileProjectItemSetting, err error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !projectNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}

	if m.Dir == "" {
		return nil, helper.Child("dir").NewValueEmptyError()
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

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newProfileProjectItemSetting(m.Name, m.Dir, m.Match, dependency, resource, matchObj), nil
}

// endregion

// region ProfileSettingModelBuilder

type ProfileSettingModelBuilder[P any] struct {
	commit   func(*profileSettingModel) P
	option   *profileOptionSettingModel
	project  *profileProjectSettingModel
	executor *executorSettingModel
	registry *registrySettingModel
	redirect *redirectSettingModel
}

func newProfileSettingModelBuilder[P any](commit func(*profileSettingModel) P) *ProfileSettingModelBuilder[P] {
	return &ProfileSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileSettingModelBuilder[P]) SetOptionSetting() *ProfileOptionSettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return newProfileOptionSettingModelBuilder(b.setOptionSettingModel)
}

func (b *ProfileSettingModelBuilder[P]) SetProjectSetting() *ProfileProjectSettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return newProfileProjectSettingModelBuilder(b.setProjectSettingModel)
}

func (b *ProfileSettingModelBuilder[P]) SetExecutorSetting() *ExecutorSettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return newExecutorSettingModelBuilder(b.setExecutorSettingModel)
}

func (b *ProfileSettingModelBuilder[P]) SetRegistrySetting() *RegistrySettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return newProfileRegistrySettingBuilder(b.setRegistrySettingModel)
}

func (b *ProfileSettingModelBuilder[P]) SetRedirectSetting() *RedirectSettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return newRedirectSettingModelBuilder(b.setRedirectSettingModel)
}

func (b *ProfileSettingModelBuilder[P]) CommitProfileSetting() P {
	return b.commit(newProfileSettingModel(b.option, b.project, b.executor, b.registry, b.redirect))
}

func (b *ProfileSettingModelBuilder[P]) setOptionSettingModel(option *profileOptionSettingModel) *ProfileSettingModelBuilder[P] {
	b.option = option
	return b
}

func (b *ProfileSettingModelBuilder[P]) setProjectSettingModel(project *profileProjectSettingModel) *ProfileSettingModelBuilder[P] {
	b.project = project
	return b
}

func (b *ProfileSettingModelBuilder[P]) setExecutorSettingModel(executor *executorSettingModel) *ProfileSettingModelBuilder[P] {
	b.executor = executor
	return b
}

func (b *ProfileSettingModelBuilder[P]) setRegistrySettingModel(registry *registrySettingModel) *ProfileSettingModelBuilder[P] {
	b.registry = registry
	return b
}

func (b *ProfileSettingModelBuilder[P]) setRedirectSettingModel(redirect *redirectSettingModel) *ProfileSettingModelBuilder[P] {
	b.redirect = redirect
	return b
}

// endregion

// region ProfileOptionSettingModelBuilder

type ProfileOptionSettingModelBuilder[P any] struct {
	commit func(*profileOptionSettingModel) P
	items  []*profileOptionItemSettingModel
}

func newProfileOptionSettingModelBuilder[P any](commit func(*profileOptionSettingModel) P) *ProfileOptionSettingModelBuilder[P] {
	return &ProfileOptionSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileOptionSettingModelBuilder[P]) AddItem(name, value, match string) *ProfileOptionSettingModelBuilder[P] {
	b.items = append(b.items, &profileOptionItemSettingModel{
		Name:  name,
		Value: value,
		Match: match,
	})
	return b
}

func (b *ProfileOptionSettingModelBuilder[P]) AddItemMap(items map[string]string) *ProfileOptionSettingModelBuilder[P] {
	for name, value := range items {
		b.AddItem(name, value, "")
	}
	return b
}

func (b *ProfileOptionSettingModelBuilder[P]) CommitOptionSetting() P {
	return b.commit(newProfileOptionSettingModel(b.items))
}

// endregion

// region ProfileProjectSettingModelBuilder

type ProfileProjectSettingModelBuilder[P any] struct {
	commit func(*profileProjectSettingModel) P
	items  []*profileProjectItemSettingModel
}

func newProfileProjectSettingModelBuilder[P any](commit func(*profileProjectSettingModel) P) *ProfileProjectSettingModelBuilder[P] {
	return &ProfileProjectSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileProjectSettingModelBuilder[P]) AddItemSetting(name, dir string) *ProfileProjectItemSettingModelBuilder[*ProfileProjectSettingModelBuilder[P]] {
	return newProfileProjectItemSettingModelBuilder(b.addItemSettingModel, name, dir)
}

func (b *ProfileProjectSettingModelBuilder[P]) CommitProjectSetting() P {
	return b.commit(newProfileProjectSettingModel(b.items))
}

func (b *ProfileProjectSettingModelBuilder[P]) addItemSettingModel(item *profileProjectItemSettingModel) *ProfileProjectSettingModelBuilder[P] {
	b.items = append(b.items, item)
	return b
}

// endregion

// region ProfileProjectItemSettingModelBuilder

type ProfileProjectItemSettingModelBuilder[P any] struct {
	commit     func(*profileProjectItemSettingModel) P
	name       string
	dir        string
	match      string
	dependency *projectDependencySettingModel
	resource   *projectResourceSettingModel
}

func newProfileProjectItemSettingModelBuilder[P any](commit func(*profileProjectItemSettingModel) P, name, dir string) *ProfileProjectItemSettingModelBuilder[P] {
	return &ProfileProjectItemSettingModelBuilder[P]{
		commit:     commit,
		name:       name,
		dir:        dir,
		dependency: newProjectDependencySettingModel(nil),
		resource:   newProjectResourceSettingModel(nil),
	}
}

func (b *ProfileProjectItemSettingModelBuilder[P]) SetMatch(match string) *ProfileProjectItemSettingModelBuilder[P] {
	b.match = match
	return b
}

func (b *ProfileProjectItemSettingModelBuilder[P]) AddDependencyItem(link, match string) *ProfileProjectItemSettingModelBuilder[P] {
	b.dependency.Items = append(b.dependency.Items, newProjectDependencyItemSettingModel(link, match))
	return b
}

func (b *ProfileProjectItemSettingModelBuilder[P]) AddResourceItem(dir string, includes, excludes []string, match string) *ProfileProjectItemSettingModelBuilder[P] {
	b.resource.Items = append(b.resource.Items, newProjectResourceItemSettingModel(dir, includes, excludes, match))
	return b
}

func (b *ProfileProjectItemSettingModelBuilder[P]) CommitItemSetting() P {
	return b.commit(newProfileProjectItemSettingModel(b.name, b.dir, b.match, b.dependency, b.resource))
}

// endregion
