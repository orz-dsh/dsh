package dsh_core

// region ProfileSettingBuilder

type ProfileSettingBuilder[P any] struct {
	commit    func(*profileSetting, error) P
	option    *profileOptionSettingModel
	project   *profileProjectSettingModel
	workspace *profileWorkspaceSettingModel
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

func (b *ProfileSettingBuilder[P]) SetWorkspaceSetting() *ProfileWorkspaceSettingBuilder[*ProfileSettingBuilder[P]] {
	return newProfileWorkspaceSettingBuilder(b.setWorkspaceSettingModel)
}

func (b *ProfileSettingBuilder[P]) CommitProfileSetting() P {
	setting, err := loadProfileSettingModel(newProfileSettingModel(b.option, b.project, b.workspace))
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

func (b *ProfileSettingBuilder[P]) setWorkspaceSettingModel(workspace *profileWorkspaceSettingModel) *ProfileSettingBuilder[P] {
	b.workspace = workspace
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
	commit  func(*profileProjectItemSettingModel) P
	name    string
	path    string
	match   string
	imports projectImportSettingModelSet
	sources projectSourceSettingModelSet
}

func newProfileProjectItemSettingBuilder[P any](commit func(*profileProjectItemSettingModel) P, name, path string) *ProfileProjectItemSettingBuilder[P] {
	return &ProfileProjectItemSettingBuilder[P]{
		commit: commit,
		name:   name,
		path:   path,
	}
}

func (b *ProfileProjectItemSettingBuilder[P]) SetMatch(match string) *ProfileProjectItemSettingBuilder[P] {
	b.match = match
	return b
}

func (b *ProfileProjectItemSettingBuilder[P]) AddImport(link, match string) *ProfileProjectItemSettingBuilder[P] {
	b.imports = append(b.imports, newProjectImportSettingModel(link, match))
	return b
}

func (b *ProfileProjectItemSettingBuilder[P]) AddSource(dir string, files []string, match string) *ProfileProjectItemSettingBuilder[P] {
	b.sources = append(b.sources, newProjectSourceSettingModel(dir, files, match))
	return b
}

func (b *ProfileProjectItemSettingBuilder[P]) CommitItemSetting() P {
	return b.commit(newProfileProjectItemSettingModel(b.name, b.path, b.match, b.imports, b.sources))
}

// endregion

// region ProfileWorkspaceSettingBuilder

type ProfileWorkspaceSettingBuilder[P any] struct {
	commit   func(*profileWorkspaceSettingModel) P
	executor *workspaceExecutorSettingModel
	import_  *workspaceImportSettingModel
}

func newProfileWorkspaceSettingBuilder[P any](commit func(*profileWorkspaceSettingModel) P) *ProfileWorkspaceSettingBuilder[P] {
	return &ProfileWorkspaceSettingBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileWorkspaceSettingBuilder[P]) SetExecutorSetting() *ProfileWorkspaceExecutorSettingBuilder[*ProfileWorkspaceSettingBuilder[P]] {
	return newProfileWorkspaceExecutorSettingBuilder(b.setExecutorSettingModel)
}

func (b *ProfileWorkspaceSettingBuilder[P]) SetImportSetting() *ProfileWorkspaceImportSettingBuilder[*ProfileWorkspaceSettingBuilder[P]] {
	return newProfileWorkspaceImportSettingBuilder(b.setImportSettingModel)
}

func (b *ProfileWorkspaceSettingBuilder[P]) CommitWorkspaceSetting() P {
	return b.commit(newProfileWorkspaceSettingModel(b.executor, b.import_))
}

func (b *ProfileWorkspaceSettingBuilder[P]) setExecutorSettingModel(executor *workspaceExecutorSettingModel) *ProfileWorkspaceSettingBuilder[P] {
	b.executor = executor
	return b
}

func (b *ProfileWorkspaceSettingBuilder[P]) setImportSettingModel(import_ *workspaceImportSettingModel) *ProfileWorkspaceSettingBuilder[P] {
	b.import_ = import_
	return b
}

// endregion

// region ProfileWorkspaceExecutorSettingBuilder

type ProfileWorkspaceExecutorSettingBuilder[P any] struct {
	commit func(*workspaceExecutorSettingModel) P
	items  []*workspaceExecutorItemSettingModel
}

func newProfileWorkspaceExecutorSettingBuilder[P any](commit func(*workspaceExecutorSettingModel) P) *ProfileWorkspaceExecutorSettingBuilder[P] {
	return &ProfileWorkspaceExecutorSettingBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileWorkspaceExecutorSettingBuilder[P]) AddItem(name, path string, exts, args []string, match string) *ProfileWorkspaceExecutorSettingBuilder[P] {
	b.items = append(b.items, newWorkspaceExecutorItemSettingModel(name, path, exts, args, match))
	return b
}

func (b *ProfileWorkspaceExecutorSettingBuilder[P]) CommitExecutorSetting() P {
	return b.commit(newWorkspaceExecutorSettingModel(b.items))
}

// endregion

// region ProfileWorkspaceImportSettingBuilder

type ProfileWorkspaceImportSettingBuilder[P any] struct {
	commit   func(*workspaceImportSettingModel) P
	registry *workspaceImportRegistrySettingModel
	redirect *workspaceImportRedirectSettingModel
}

func newProfileWorkspaceImportSettingBuilder[P any](commit func(*workspaceImportSettingModel) P) *ProfileWorkspaceImportSettingBuilder[P] {
	return &ProfileWorkspaceImportSettingBuilder[P]{
		commit:   commit,
		registry: &workspaceImportRegistrySettingModel{},
		redirect: &workspaceImportRedirectSettingModel{},
	}
}

func (b *ProfileWorkspaceImportSettingBuilder[P]) AddRegistryItem(name, link, match string) *ProfileWorkspaceImportSettingBuilder[P] {
	b.registry.Items = append(b.registry.Items, newWorkspaceImportRegistryItemSettingModel(name, link, match))
	return b
}

func (b *ProfileWorkspaceImportSettingBuilder[P]) AddRedirectItem(regex, link, match string) *ProfileWorkspaceImportSettingBuilder[P] {
	b.redirect.Items = append(b.redirect.Items, newWorkspaceImportRedirectItemSettingModel(regex, link, match))
	return b
}

func (b *ProfileWorkspaceImportSettingBuilder[P]) CommitImportSetting() P {
	return b.commit(newWorkspaceImportSettingModel(b.registry, b.redirect))
}

// endregion
