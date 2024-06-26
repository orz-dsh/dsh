package dsh_core

// region ProfileSettingBuilder

type ProfileSettingBuilder[P any] struct {
	commit   func(*profileSetting, error) P
	option   *profileOptionSettingModel
	project  *profileProjectSettingModel
	executor *workspaceExecutorSettingModel
	registry *workspaceRegistrySettingModel
	redirect *workspaceRedirectSettingModel
}

func newProfileSettingBuilder[P any](commit func(*profileSetting, error) P) *ProfileSettingBuilder[P] {
	return &ProfileSettingBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileSettingBuilder[P]) SetOptionSetting() *ProfileOptionSettingBuilder[*ProfileSettingBuilder[P]] {
	return newProfileOptionSettingBuilder(b.setOptionSettingModel)
}

func (b *ProfileSettingBuilder[P]) SetProjectSetting() *ProfileProjectSettingBuilder[*ProfileSettingBuilder[P]] {
	return newProfileProjectSettingBuilder(b.setProjectSettingModel)
}

func (b *ProfileSettingBuilder[P]) SetExecutorSetting() *ProfileExecutorSettingBuilder[*ProfileSettingBuilder[P]] {
	return newProfileExecutorSettingBuilder(b.setExecutorSettingModel)
}

func (b *ProfileSettingBuilder[P]) SetRegistrySetting() *ProfileRegistrySettingBuilder[*ProfileSettingBuilder[P]] {
	return newProfileRegistrySettingBuilder(b.setRegistrySettingModel)
}

func (b *ProfileSettingBuilder[P]) SetRedirectSetting() *ProfileRedirectSettingBuilder[*ProfileSettingBuilder[P]] {
	return newProfileRedirectSettingBuilder(b.setRedirectSettingModel)
}

func (b *ProfileSettingBuilder[P]) CommitProfileSetting() P {
	setting, err := loadProfileSettingModel(newProfileSettingModel(b.option, b.project, b.executor, b.registry, b.redirect))
	return b.commit(setting, err)
}

func (b *ProfileSettingBuilder[P]) setOptionSettingModel(option *profileOptionSettingModel) *ProfileSettingBuilder[P] {
	b.option = option
	return b
}

func (b *ProfileSettingBuilder[P]) setProjectSettingModel(project *profileProjectSettingModel) *ProfileSettingBuilder[P] {
	b.project = project
	return b
}

func (b *ProfileSettingBuilder[P]) setExecutorSettingModel(executor *workspaceExecutorSettingModel) *ProfileSettingBuilder[P] {
	b.executor = executor
	return b
}

func (b *ProfileSettingBuilder[P]) setRegistrySettingModel(registry *workspaceRegistrySettingModel) *ProfileSettingBuilder[P] {
	b.registry = registry
	return b
}

func (b *ProfileSettingBuilder[P]) setRedirectSettingModel(redirect *workspaceRedirectSettingModel) *ProfileSettingBuilder[P] {
	b.redirect = redirect
	return b
}

// endregion

// region ProfileOptionSettingBuilder

type ProfileOptionSettingBuilder[P any] struct {
	commit func(*profileOptionSettingModel) P
	items  []*profileOptionItemSettingModel
}

func newProfileOptionSettingBuilder[P any](commit func(*profileOptionSettingModel) P) *ProfileOptionSettingBuilder[P] {
	return &ProfileOptionSettingBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileOptionSettingBuilder[P]) AddItem(name, value, match string) *ProfileOptionSettingBuilder[P] {
	b.items = append(b.items, &profileOptionItemSettingModel{
		Name:  name,
		Value: value,
		Match: match,
	})
	return b
}

func (b *ProfileOptionSettingBuilder[P]) AddItemMap(items map[string]string) *ProfileOptionSettingBuilder[P] {
	for name, value := range items {
		b.AddItem(name, value, "")
	}
	return b
}

func (b *ProfileOptionSettingBuilder[P]) CommitOptionSetting() P {
	return b.commit(newProfileOptionSettingModel(b.items))
}

// endregion

// region ProfileProjectSettingBuilder

type ProfileProjectSettingBuilder[P any] struct {
	commit func(*profileProjectSettingModel) P
	items  []*profileProjectItemSettingModel
}

func newProfileProjectSettingBuilder[P any](commit func(*profileProjectSettingModel) P) *ProfileProjectSettingBuilder[P] {
	return &ProfileProjectSettingBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileProjectSettingBuilder[P]) AddItemSetting(name, path string) *ProfileProjectItemSettingBuilder[*ProfileProjectSettingBuilder[P]] {
	return newProfileProjectItemSettingBuilder(b.addItemSettingModel, name, path)
}

func (b *ProfileProjectSettingBuilder[P]) CommitProjectSetting() P {
	return b.commit(newProfileProjectSettingModel(b.items))
}

func (b *ProfileProjectSettingBuilder[P]) addItemSettingModel(item *profileProjectItemSettingModel) *ProfileProjectSettingBuilder[P] {
	b.items = append(b.items, item)
	return b
}

// endregion

// region ProfileProjectItemSettingBuilder

type ProfileProjectItemSettingBuilder[P any] struct {
	commit     func(*profileProjectItemSettingModel) P
	name       string
	path       string
	match      string
	dependency *projectDependencySettingModel
	resource   *projectResourceSettingModel
}

func newProfileProjectItemSettingBuilder[P any](commit func(*profileProjectItemSettingModel) P, name, path string) *ProfileProjectItemSettingBuilder[P] {
	return &ProfileProjectItemSettingBuilder[P]{
		commit:     commit,
		name:       name,
		path:       path,
		dependency: newProjectDependencySettingModel(nil),
		resource:   newProjectResourceSettingModel(nil),
	}
}

func (b *ProfileProjectItemSettingBuilder[P]) SetMatch(match string) *ProfileProjectItemSettingBuilder[P] {
	b.match = match
	return b
}

func (b *ProfileProjectItemSettingBuilder[P]) AddDependencyItem(link, match string) *ProfileProjectItemSettingBuilder[P] {
	b.dependency.Items = append(b.dependency.Items, newProjectDependencyItemSettingModel(link, match))
	return b
}

func (b *ProfileProjectItemSettingBuilder[P]) AddResourceItem(dir string, includes, excludes []string, match string) *ProfileProjectItemSettingBuilder[P] {
	b.resource.Items = append(b.resource.Items, newProjectResourceItemSettingModel(dir, includes, excludes, match))
	return b
}

func (b *ProfileProjectItemSettingBuilder[P]) CommitItemSetting() P {
	return b.commit(newProfileProjectItemSettingModel(b.name, b.path, b.match, b.dependency, b.resource))
}

// endregion

// region ProfileExecutorSettingBuilder

type ProfileExecutorSettingBuilder[P any] struct {
	commit func(*workspaceExecutorSettingModel) P
	items  []*workspaceExecutorItemSettingModel
}

func newProfileExecutorSettingBuilder[P any](commit func(*workspaceExecutorSettingModel) P) *ProfileExecutorSettingBuilder[P] {
	return &ProfileExecutorSettingBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileExecutorSettingBuilder[P]) AddItem(name, path string, exts, args []string, match string) *ProfileExecutorSettingBuilder[P] {
	b.items = append(b.items, newWorkspaceExecutorItemSettingModel(name, path, exts, args, match))
	return b
}

func (b *ProfileExecutorSettingBuilder[P]) CommitExecutorSetting() P {
	return b.commit(newWorkspaceExecutorSettingModel(b.items))
}

// endregion

// region ProfileRegistrySettingBuilder

type ProfileRegistrySettingBuilder[P any] struct {
	commit func(*workspaceRegistrySettingModel) P
	items  []*workspaceRegistryItemSettingModel
}

func newProfileRegistrySettingBuilder[P any](commit func(*workspaceRegistrySettingModel) P) *ProfileRegistrySettingBuilder[P] {
	return &ProfileRegistrySettingBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileRegistrySettingBuilder[P]) AddItem(name, link, match string) *ProfileRegistrySettingBuilder[P] {
	b.items = append(b.items, newWorkspaceRegistryItemSettingModel(name, link, match))
	return b
}

func (b *ProfileRegistrySettingBuilder[P]) CommitRegistrySetting() P {
	return b.commit(newWorkspaceRegistrySettingModel(b.items))
}

// endregion

// region ProfileRedirectSettingBuilder

type ProfileRedirectSettingBuilder[P any] struct {
	commit func(*workspaceRedirectSettingModel) P
	items  []*workspaceRedirectItemSettingModel
}

func newProfileRedirectSettingBuilder[P any](commit func(*workspaceRedirectSettingModel) P) *ProfileRedirectSettingBuilder[P] {
	return &ProfileRedirectSettingBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileRedirectSettingBuilder[P]) AddItem(regex, link, match string) *ProfileRedirectSettingBuilder[P] {
	b.items = append(b.items, newWorkspaceRedirectItemSettingModel(regex, link, match))
	return b
}

func (b *ProfileRedirectSettingBuilder[P]) CommitRedirectSetting() P {
	return b.commit(newWorkspaceRedirectSettingModel(b.items))
}

// endregion
