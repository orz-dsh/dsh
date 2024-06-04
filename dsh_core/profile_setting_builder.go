package dsh_core

// region ProfileSettingBuilder

type ProfileSettingBuilder struct {
	option    *profileOptionSettingModel
	project   *profileProjectSettingModel
	workspace *profileWorkspaceSettingModel
}

func NewProfileSettingBuilder() *ProfileSettingBuilder {
	return &ProfileSettingBuilder{}
}

func (b *ProfileSettingBuilder) Option() *ProfileOptionSettingBuilder {
	return newProfileOptionSettingBuilder(b)
}

func (b *ProfileSettingBuilder) Project() *ProfileProjectSettingBuilder {
	return newProfileProjectSettingBuilder(b)
}

func (b *ProfileSettingBuilder) Workspace() *ProfileWorkspaceSettingBuilder {
	return newProfileWorkspaceSettingBuilder(b)
}

func (b *ProfileSettingBuilder) setOption(option *profileOptionSettingModel) *ProfileSettingBuilder {
	b.option = option
	return b
}

func (b *ProfileSettingBuilder) setProject(project *profileProjectSettingModel) *ProfileSettingBuilder {
	b.project = project
	return b
}

func (b *ProfileSettingBuilder) setWorkspace(workspace *profileWorkspaceSettingModel) *ProfileSettingBuilder {
	b.workspace = workspace
	return b
}

func (b *ProfileSettingBuilder) buildModel() *profileSettingModel {
	return newProfileSettingModel(b.option, b.project, b.workspace)
}

// endregion

// region ProfileOptionSettingBuilder

type ProfileOptionSettingBuilder struct {
	parent *ProfileSettingBuilder
	items  []*profileOptionItemSettingModel
}

func newProfileOptionSettingBuilder(parent *ProfileSettingBuilder) *ProfileOptionSettingBuilder {
	return &ProfileOptionSettingBuilder{
		parent: parent,
	}
}

func (b *ProfileOptionSettingBuilder) AddItem(name, value, match string) *ProfileOptionSettingBuilder {
	b.items = append(b.items, &profileOptionItemSettingModel{
		Name:  name,
		Value: value,
		Match: match,
	})
	return b
}

func (b *ProfileOptionSettingBuilder) AddItemMap(items map[string]string) *ProfileOptionSettingBuilder {
	for name, value := range items {
		b.AddItem(name, value, "")
	}
	return b
}

func (b *ProfileOptionSettingBuilder) Commit() *ProfileSettingBuilder {
	return b.parent.setOption(b.buildModel())
}

func (b *ProfileOptionSettingBuilder) buildModel() *profileOptionSettingModel {
	return newProfileOptionSettingModel(b.items)
}

// endregion

// region ProfileProjectSettingBuilder

type ProfileProjectSettingBuilder struct {
	parent *ProfileSettingBuilder
	items  []*profileProjectItemSettingModel
}

func newProfileProjectSettingBuilder(parent *ProfileSettingBuilder) *ProfileProjectSettingBuilder {
	return &ProfileProjectSettingBuilder{
		parent: parent,
	}
}

func (b *ProfileProjectSettingBuilder) Item(name, path string) *ProfileProjectItemSettingBuilder {
	return newProfileProjectItemSettingBuilder(b, name, path)
}

func (b *ProfileProjectSettingBuilder) Commit() *ProfileSettingBuilder {
	return b.parent.setProject(b.buildModel())
}

func (b *ProfileProjectSettingBuilder) addItem(item *profileProjectItemSettingModel) *ProfileProjectSettingBuilder {
	b.items = append(b.items, item)
	return b
}

func (b *ProfileProjectSettingBuilder) buildModel() *profileProjectSettingModel {
	return newProfileProjectSettingModel(b.items)
}

// endregion

// region ProfileProjectItemSettingBuilder

type ProfileProjectItemSettingBuilder struct {
	parent *ProfileProjectSettingBuilder
	name   string
	path   string
	match  string
	script *projectScriptSettingModel
	config *projectConfigSettingModel
}

func newProfileProjectItemSettingBuilder(parent *ProfileProjectSettingBuilder, name, path string) *ProfileProjectItemSettingBuilder {
	return &ProfileProjectItemSettingBuilder{
		parent: parent,
		name:   name,
		path:   path,
		script: &projectScriptSettingModel{},
		config: &projectConfigSettingModel{},
	}
}

func (b *ProfileProjectItemSettingBuilder) setMatch(match string) *ProfileProjectItemSettingBuilder {
	b.match = match
	return b
}

func (b *ProfileProjectItemSettingBuilder) AddScriptSource(dir string, files []string, match string) *ProfileProjectItemSettingBuilder {
	b.script.Sources = append(b.script.Sources, newProjectSourceSettingModel(dir, files, match))
	return b
}

func (b *ProfileProjectItemSettingBuilder) AddScriptImport(link, match string) *ProfileProjectItemSettingBuilder {
	b.script.Imports = append(b.script.Imports, newProjectImportSettingModel(link, match))
	return b
}

func (b *ProfileProjectItemSettingBuilder) AddConfigSource(dir string, files []string, match string) *ProfileProjectItemSettingBuilder {
	b.config.Sources = append(b.config.Sources, newProjectSourceSettingModel(dir, files, match))
	return b
}

func (b *ProfileProjectItemSettingBuilder) AddConfigImport(link, match string) *ProfileProjectItemSettingBuilder {
	b.config.Imports = append(b.config.Imports, newProjectImportSettingModel(link, match))
	return b
}

func (b *ProfileProjectItemSettingBuilder) Commit() *ProfileProjectSettingBuilder {
	return b.parent.addItem(b.buildModel())
}

func (b *ProfileProjectItemSettingBuilder) buildModel() *profileProjectItemSettingModel {
	return newProfileProjectItemSettingModel(b.name, b.path, b.match, b.script, b.config)
}

// endregion

// region ProfileWorkspaceSettingBuilder

type ProfileWorkspaceSettingBuilder struct {
	parent  *ProfileSettingBuilder
	shell   *workspaceShellSettingModel
	import_ *workspaceImportSettingModel
}

func newProfileWorkspaceSettingBuilder(parent *ProfileSettingBuilder) *ProfileWorkspaceSettingBuilder {
	return &ProfileWorkspaceSettingBuilder{
		parent: parent,
	}
}

func (b *ProfileWorkspaceSettingBuilder) Shell() *ProfileWorkspaceShellSettingBuilder {
	return newProfileWorkspaceShellSettingBuilder(b)
}

func (b *ProfileWorkspaceSettingBuilder) Import() *ProfileWorkspaceImportSettingBuilder {
	return newProfileWorkspaceImportSettingBuilder(b)
}

func (b *ProfileWorkspaceSettingBuilder) Commit() *ProfileSettingBuilder {
	return b.parent.setWorkspace(b.buildModel())
}

func (b *ProfileWorkspaceSettingBuilder) setShell(shell *workspaceShellSettingModel) *ProfileWorkspaceSettingBuilder {
	b.shell = shell
	return b
}

func (b *ProfileWorkspaceSettingBuilder) setImport(import_ *workspaceImportSettingModel) *ProfileWorkspaceSettingBuilder {
	b.import_ = import_
	return b
}

func (b *ProfileWorkspaceSettingBuilder) buildModel() *profileWorkspaceSettingModel {
	return newProfileWorkspaceSettingModel(b.shell, b.import_)
}

// endregion

// region ProfileWorkspaceShellSettingBuilder

type ProfileWorkspaceShellSettingBuilder struct {
	parent *ProfileWorkspaceSettingBuilder
	items  []*workspaceShellItemSettingModel
}

func newProfileWorkspaceShellSettingBuilder(parent *ProfileWorkspaceSettingBuilder) *ProfileWorkspaceShellSettingBuilder {
	return &ProfileWorkspaceShellSettingBuilder{
		parent: parent,
	}
}

func (b *ProfileWorkspaceShellSettingBuilder) AddItem(name, path string, exts, args []string, match string) *ProfileWorkspaceShellSettingBuilder {
	b.items = append(b.items, newWorkspaceShellItemSettingModel(name, path, exts, args, match))
	return b
}

func (b *ProfileWorkspaceShellSettingBuilder) Commit() *ProfileWorkspaceSettingBuilder {
	return b.parent.setShell(b.buildModel())
}

func (b *ProfileWorkspaceShellSettingBuilder) buildModel() *workspaceShellSettingModel {
	return newWorkspaceShellSettingModel(b.items)
}

// endregion

// region ProfileWorkspaceImportSettingBuilder

type ProfileWorkspaceImportSettingBuilder struct {
	parent   *ProfileWorkspaceSettingBuilder
	registry *workspaceImportRegistrySettingModel
	redirect *workspaceImportRedirectSettingModel
}

func newProfileWorkspaceImportSettingBuilder(parent *ProfileWorkspaceSettingBuilder) *ProfileWorkspaceImportSettingBuilder {
	return &ProfileWorkspaceImportSettingBuilder{
		parent:   parent,
		registry: &workspaceImportRegistrySettingModel{},
		redirect: &workspaceImportRedirectSettingModel{},
	}
}

func (b *ProfileWorkspaceImportSettingBuilder) AddRegistryItem(name, link, match string) *ProfileWorkspaceImportSettingBuilder {
	b.registry.Items = append(b.registry.Items, newWorkspaceImportRegistryItemSettingModel(name, link, match))
	return b
}

func (b *ProfileWorkspaceImportSettingBuilder) AddRedirectItem(regex, link, match string) *ProfileWorkspaceImportSettingBuilder {
	b.redirect.Items = append(b.redirect.Items, newWorkspaceImportRedirectItemSettingModel(regex, link, match))
	return b
}

func (b *ProfileWorkspaceImportSettingBuilder) Commit() *ProfileWorkspaceSettingBuilder {
	return b.parent.setImport(b.buildModel())
}

func (b *ProfileWorkspaceImportSettingBuilder) buildModel() *workspaceImportSettingModel {
	return newWorkspaceImportSettingModel(b.registry, b.redirect)
}

// endregion
