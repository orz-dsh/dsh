package dsh_core

// region ProfileSettingBuilder

type ProfileSettingBuilder struct {
	option    *profileOptionSettingModel
	project   *profileProjectSettingModel
	workspace *profileWorkspaceSettingModel
}

func NewProfileSettingBuilder() *ProfileSettingBuilder {
	return &ProfileSettingBuilder{
		option:    &profileOptionSettingModel{},
		project:   &profileProjectSettingModel{},
		workspace: &profileWorkspaceSettingModel{},
	}
}

func (b *ProfileSettingBuilder) AddOptionItem(name string, value string, match string) *ProfileSettingBuilder {
	b.option.Items = append(b.option.Items, &profileOptionItemSettingModel{
		Name:  name,
		Value: value,
		Match: match,
	})
	return b
}

func (b *ProfileSettingBuilder) AddOptionItemMap(items map[string]string) *ProfileSettingBuilder {
	for name, value := range items {
		b.AddOptionItem(name, value, "")
	}
	return b
}

func (b *ProfileSettingBuilder) AddProjectItem(builder *ProfileProjectItemSettingBuilder) *ProfileSettingBuilder {
	b.project.Items = append(b.project.Items, builder.buildModel())
	return b
}

func (b *ProfileSettingBuilder) buildModel() *profileSettingModel {
	return &profileSettingModel{
		Option:    b.option,
		Project:   b.project,
		Workspace: b.workspace,
	}
}

// endregion

// region ProfileProjectItemSettingBuilder

type ProfileProjectItemSettingBuilder struct {
	name   string
	path   string
	match  string
	script *projectScriptSettingModel
	config *projectConfigSettingModel
}

func NewProfileProjectItemSettingBuilder(name, path string) *ProfileProjectItemSettingBuilder {
	return &ProfileProjectItemSettingBuilder{
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

func (b *ProfileProjectItemSettingBuilder) buildModel() *profileProjectItemSettingModel {
	return newProfileProjectItemSettingModel(b.name, b.path, b.match, b.script, b.config)
}

// endregion

// region ProfileWorkspaceSettingBuilder

type ProfileWorkspaceSettingBuilder struct {
	shell   *workspaceShellSettingModel
	import_ *workspaceImportSettingModel
}

func NewProfileWorkspaceSettingBuilder() *ProfileWorkspaceSettingBuilder {
	return &ProfileWorkspaceSettingBuilder{
		shell: &workspaceShellSettingModel{},
		import_: &workspaceImportSettingModel{
			Registry: &workspaceImportRegistrySettingModel{},
			Redirect: &workspaceImportRedirectSettingModel{},
		},
	}
}

func (b *ProfileWorkspaceSettingBuilder) AddShellItem(name, path string, exts, args []string, match string) *ProfileWorkspaceSettingBuilder {
	b.shell.Items = append(b.shell.Items, newWorkspaceShellItemSettingModel(name, path, exts, args, match))
	return b
}

func (b *ProfileWorkspaceSettingBuilder) AddImportRegistryItem(name, link, match string) *ProfileWorkspaceSettingBuilder {
	b.import_.Registry.Items = append(b.import_.Registry.Items, newWorkspaceImportRegistryItemSettingModel(name, link, match))
	return b
}

func (b *ProfileWorkspaceSettingBuilder) AddImportRedirectItem(regex, link, match string) *ProfileWorkspaceSettingBuilder {
	b.import_.Redirect.Items = append(b.import_.Redirect.Items, newWorkspaceImportRedirectItemSettingModel(regex, link, match))
	return b
}

func (b *ProfileWorkspaceSettingBuilder) buildModel() *profileWorkspaceSettingModel {
	return newProfileWorkspaceSettingModel(b.shell, b.import_)
}

// endregion
