package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"regexp"
)

// region manifest

type ProfileManifest struct {
	Option       *ProfileManifestOption
	Workspace    *ProfileManifestWorkspace
	Project      *ProfileManifestProject
	manifestPath string
	manifestType manifestMetadataType
}

func loadProfileManifest(path string) (*ProfileManifest, error) {
	manifest := &ProfileManifest{
		Option:    NewProfileManifestOption(nil),
		Workspace: NewProfileManifestWorkspace(nil, nil),
		Project:   NewProfileManifestProject(nil),
	}

	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, errW(err, "load profile manifest error",
				reason("get abs-path error"),
				kv("path", path),
			)
		}
		path = absPath
	}

	if path != "" {
		metadata, err := loadManifestFromFile(path, "", manifest)
		if err != nil {
			return nil, errW(err, "load profile manifest error",
				reason("load manifest from file error"),
				kv("path", path),
			)
		}
		manifest.manifestPath = metadata.ManifestPath
		manifest.manifestType = metadata.ManifestType
	}

	if err := manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func MakeProfileManifest(option *ProfileManifestOption, workspace *ProfileManifestWorkspace, project *ProfileManifestProject) (*ProfileManifest, error) {
	if option == nil {
		option = NewProfileManifestOption(nil)
	}
	if workspace == nil {
		workspace = NewProfileManifestWorkspace(nil, nil)
	}
	if project == nil {
		project = NewProfileManifestProject(nil)
	}
	manifest := &ProfileManifest{
		Option:    option,
		Workspace: workspace,
		Project:   project,
	}
	if err := manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *ProfileManifest) DescExtraKeyValues() KVS {
	return KVS{
		kv("manifestPath", m.manifestPath),
		kv("manifestType", m.manifestType),
	}
}

func (m *ProfileManifest) init() (err error) {
	if err = m.Option.init(m); err != nil {
		return err
	}
	if err = m.Workspace.init(m); err != nil {
		return err
	}
	if err = m.Project.init(m); err != nil {
		return err
	}

	return nil
}

// endregion

// region option

type ProfileManifestOption struct {
	Items    []*ProfileManifestOptionItem
	entities profileOptionSpecifyEntitySet
}

type ProfileManifestOptionItem struct {
	Name  string
	Value string
	Match string
}

func NewProfileManifestOption(items map[string]string) *ProfileManifestOption {
	var optionItems []*ProfileManifestOptionItem
	for k, v := range items {
		optionItems = append(optionItems, &ProfileManifestOptionItem{
			Name:  k,
			Value: v,
		})
	}
	return &ProfileManifestOption{
		Items: optionItems,
	}
}

func (o *ProfileManifestOption) init(manifest *ProfileManifest) (err error) {
	entities := profileOptionSpecifyEntitySet{}
	for i := 0; i < len(o.Items); i++ {
		item := o.Items[i]
		if item.Name == "" {
			return errN("profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].name", i)),
			)
		}
		var matchObj *EvalExpr
		if item.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "profile manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].match", i)),
					kv("value", item.Match),
				)
			}
		}
		entities = append(entities, newProfileOptionSpecifyEntity(item.Name, item.Value, item.Match, matchObj))
	}

	o.entities = entities
	return nil
}

// endregion

// region workspace

type ProfileManifestWorkspace struct {
	Shell  *ProfileManifestWorkspaceShell
	Import *ProfileManifestWorkspaceImport
}

func NewProfileManifestWorkspace(shell *ProfileManifestWorkspaceShell, imp *ProfileManifestWorkspaceImport) *ProfileManifestWorkspace {
	if shell == nil {
		shell = &ProfileManifestWorkspaceShell{}
	}
	if imp == nil {
		imp = &ProfileManifestWorkspaceImport{}
	}
	return &ProfileManifestWorkspace{
		Shell:  shell,
		Import: imp,
	}
}

func (w *ProfileManifestWorkspace) init(manifest *ProfileManifest) (err error) {
	if err = w.Shell.init(manifest); err != nil {
		return err
	}
	if err = w.Import.init(manifest); err != nil {
		return err
	}
	return nil
}

// endregion

// region workspace shell

type ProfileManifestWorkspaceShell struct {
	Items    []*ProfileManifestWorkspaceShellItem
	entities workspaceShellEntitySet
}

type ProfileManifestWorkspaceShellItem struct {
	Name  string
	Path  string
	Exts  []string
	Args  []string
	Match string
}

func (s *ProfileManifestWorkspaceShell) init(manifest *ProfileManifest) (err error) {
	entities := workspaceShellEntitySet{}
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if item.Name == "" {
			return errN("profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.shell.items[%d].name", i)),
			)
		}
		if item.Path != "" && !dsh_utils.IsFileExists(item.Path) {
			return errN("profile manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.shell.items[%d].path", i)),
				kv("value", item.Path),
			)
		}
		for j := 0; j < len(item.Exts); j++ {
			if item.Exts[j] == "" {
				return errN("profile manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].exts[%d]", i, j)),
				)
			}
		}
		for j := 0; j < len(item.Args); j++ {
			if item.Args[j] == "" {
				return errN("profile manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].args[%d]", i, j)),
				)
			}
		}
		var matchObj *EvalExpr
		if item.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "profile manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].match", i)),
					kv("value", item.Match),
				)
			}
		}
		entities[item.Name] = append(entities[item.Name], newWorkspaceShellEntity(item.Name, item.Path, item.Exts, item.Args, item.Match, matchObj))
	}

	s.entities = entities
	return nil
}

// endregion

// region workspace import

type ProfileManifestWorkspaceImport struct {
	Registries       []*ProfileManifestImportRegistry
	Redirects        []*ProfileManifestImportRedirect
	registryEntities workspaceImportRegistryEntitySet
	redirectEntities workspaceImportRedirectEntitySet
}

type ProfileManifestImportRegistry struct {
	Name  string
	Link  string
	Match string
}

type ProfileManifestImportRedirect struct {
	Regex string
	Link  string
	Match string
}

func (imp *ProfileManifestWorkspaceImport) init(manifest *ProfileManifest) (err error) {
	registryEntities := workspaceImportRegistryEntitySet{}
	for i := 0; i < len(imp.Registries); i++ {
		registry := imp.Registries[i]
		if registry.Name == "" {
			return errN("profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.registries[%d].name", i)),
			)
		}

		if registry.Link == "" {
			return errN("profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.registries[%d].link", i)),
			)
		}
		// TODO: check link template

		var matchObj *EvalExpr
		if registry.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(registry.Match)
			if err != nil {
				return errW(err, "profile manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.import.registries[%d].match", i)),
					kv("value", registry.Match),
				)
			}
		}
		registryEntities[registry.Name] = append(registryEntities[registry.Name], newWorkspaceImportRegistryEntity(registry.Name, registry.Link, registry.Match, matchObj))
	}

	redirectEntities := workspaceImportRedirectEntitySet{}
	for i := 0; i < len(imp.Redirects); i++ {
		redirect := imp.Redirects[i]
		if redirect.Regex == "" {
			return errN("profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.redirects[%d].regex", i)),
			)
		}
		regexObj, err := regexp.Compile(redirect.Regex)
		if err != nil {
			return errW(err, "profile manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.redirects[%d].regex", i)),
				kv("value", redirect.Regex),
			)
		}

		if redirect.Link == "" {
			return errN("profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.redirects[%d].link", i)),
			)
		}
		// TODO: check link template

		var matchObj *EvalExpr
		if redirect.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(redirect.Match)
			if err != nil {
				return errW(err, "profile manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.import.redirects[%d].match", i)),
					kv("value", redirect.Match),
				)
			}
		}
		redirectEntities = append(redirectEntities, newWorkspaceImportRedirectEntity(redirect.Regex, redirect.Link, redirect.Match, regexObj, matchObj))
	}

	imp.registryEntities = registryEntities
	imp.redirectEntities = redirectEntities
	return nil
}

// endregion

// region project

type ProfileManifestProject struct {
	Items    []*ProfileManifestProjectItem
	entities profileProjectEntitySet
}

type ProfileManifestProjectItem struct {
	Name   string
	Path   string
	Match  string
	Script *ProfileManifestProjectScript
	Config *ProfileManifestProjectConfig
}

func NewProfileManifestProject(items []*ProfileManifestProjectItem) *ProfileManifestProject {
	return &ProfileManifestProject{
		Items: items,
	}
}

func NewProfileManifestProjectItem(script *ProfileManifestProjectScript, config *ProfileManifestProjectConfig) *ProfileManifestProjectItem {
	if script == nil {
		script = &ProfileManifestProjectScript{}
	}
	if config == nil {
		config = &ProfileManifestProjectConfig{}
	}
	return &ProfileManifestProjectItem{
		Script: script,
		Config: config,
	}
}

func (p *ProfileManifestProject) init(manifest *ProfileManifest) error {
	entities := profileProjectEntitySet{}
	for i := 0; i < len(p.Items); i++ {
		if entity, err := p.Items[i].init(manifest, i); err != nil {
			return err
		} else {
			entities = append(entities, entity)
		}
	}

	p.entities = entities
	return nil
}

func (i *ProfileManifestProjectItem) init(manifest *ProfileManifest, itemIndex int) (entity *profileProjectEntity, err error) {
	if i.Name == "" {
		return nil, errN("profile manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.items[%d].name", itemIndex)),
		)
	}
	if checked := projectNameCheckRegex.MatchString(i.Name); !checked {
		return nil, errN("profile manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.items[%d].name", itemIndex)),
			kv("value", i.Name),
		)
	}

	if i.Path == "" {
		return nil, errN("profile manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.items[%d].path", itemIndex)),
		)
	}

	var scriptSources projectSourceEntitySet
	var scriptImports projectImportEntitySet
	if i.Script != nil {
		if scriptSources, scriptImports, err = i.Script.init(manifest, itemIndex); err != nil {
			return nil, err
		}
	}

	var configSources projectSourceEntitySet
	var configImports projectImportEntitySet
	if i.Config != nil {
		if configSources, configImports, err = i.Config.init(manifest, itemIndex); err != nil {
			return nil, err
		}
	}

	var matchObj *EvalExpr
	if i.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(i.Match)
		if err != nil {
			return nil, errW(err, "profile manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("project.items[%d].match", itemIndex)),
				kv("value", i.Match),
			)
		}
	}

	entity = newProfileProjectEntity(i.Name, i.Path, i.Match, scriptSources, scriptImports, configSources, configImports, matchObj)
	return entity, nil
}

// endregion

// region project script

type ProfileManifestProjectScript struct {
	Sources []*ProfileManifestProjectSource
	Imports []*ProfileManifestProjectImport
}

func NewProfileManifestProjectScript(sources []*ProfileManifestProjectSource, imports []*ProfileManifestProjectImport) *ProfileManifestProjectScript {
	return &ProfileManifestProjectScript{
		Sources: sources,
		Imports: imports,
	}
}

func (s *ProfileManifestProjectScript) init(manifest *ProfileManifest, itemIndex int) (projectSourceEntitySet, projectImportEntitySet, error) {
	sources := projectSourceEntitySet{}
	for sourceIndex := 0; sourceIndex < len(s.Sources); sourceIndex++ {
		src := s.Sources[sourceIndex]
		if entity, err := src.init(manifest, "script", itemIndex, sourceIndex); err != nil {
			return nil, nil, err
		} else {
			sources = append(sources, entity)
		}
	}

	imports := projectImportEntitySet{}
	for importIndex := 0; importIndex < len(s.Imports); importIndex++ {
		imp := s.Imports[importIndex]
		if entity, err := imp.init(manifest, "script", itemIndex, importIndex); err != nil {
			return nil, nil, err
		} else {
			imports = append(imports, entity)
		}
	}

	return sources, imports, nil
}

// endregion

// region project config

type ProfileManifestProjectConfig struct {
	Sources []*ProfileManifestProjectSource
	Imports []*ProfileManifestProjectImport
}

func NewProfileManifestProjectConfig(sources []*ProfileManifestProjectSource, imports []*ProfileManifestProjectImport) *ProfileManifestProjectConfig {
	return &ProfileManifestProjectConfig{
		Sources: sources,
		Imports: imports,
	}
}

func (c *ProfileManifestProjectConfig) init(manifest *ProfileManifest, itemIndex int) (projectSourceEntitySet, projectImportEntitySet, error) {
	sources := projectSourceEntitySet{}
	for sourceIndex := 0; sourceIndex < len(c.Sources); sourceIndex++ {
		src := c.Sources[sourceIndex]
		if entity, err := src.init(manifest, "config", itemIndex, sourceIndex); err != nil {
			return nil, nil, err
		} else {
			sources = append(sources, entity)
		}
	}

	imports := projectImportEntitySet{}
	for importIndex := 0; importIndex < len(c.Imports); importIndex++ {
		imp := c.Imports[importIndex]
		if entity, err := imp.init(manifest, "config", itemIndex, importIndex); err != nil {
			return nil, nil, err
		} else {
			imports = append(imports, entity)
		}
	}

	return sources, imports, nil
}

// endregion

// region project source

type ProfileManifestProjectSource struct {
	Dir   string
	Files []string
	Match string
}

func NewProfileManifestProjectSource(dir string, files []string, match string) *ProfileManifestProjectSource {
	return &ProfileManifestProjectSource{
		Dir:   dir,
		Files: files,
		Match: match,
	}
}

func (s *ProfileManifestProjectSource) init(manifest *ProfileManifest, scope string, itemIndex, sourceIndex int) (entity *projectSourceEntity, err error) {
	if s.Dir == "" {
		return nil, errN("profile manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.items[%d].%s.sources[%d].dir", itemIndex, scope, sourceIndex)),
		)
	}

	var matchObj *EvalExpr
	if s.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(s.Match)
		if err != nil {
			return nil, errW(err, "profile manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("project.items[%d].%s.sources[%d].match", itemIndex, scope, sourceIndex)),
				kv("value", s.Match),
			)
		}
	}

	return newProjectSourceEntity(s.Dir, s.Files, s.Match, matchObj), nil
}

// endregion

// region project import

type ProfileManifestProjectImport struct {
	Link  string
	Match string
}

func NewProfileManifestProjectImport(link string, match string) *ProfileManifestProjectImport {
	return &ProfileManifestProjectImport{
		Link:  link,
		Match: match,
	}
}

func (i *ProfileManifestProjectImport) init(manifest *ProfileManifest, scope string, itemIndex, importIndex int) (entity *projectImportEntity, err error) {
	if i.Link == "" {
		return nil, errN("profile manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.items[%d].%s.imports[%d].link", itemIndex, scope, importIndex)),
		)
	}
	linkObj, err := parseProjectLink(i.Link)
	if err != nil {
		return nil, errW(err, "profile manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.items[%d].%s.imports[%d].link", itemIndex, scope, importIndex)),
			kv("value", i.Link),
		)
	}

	var matchObj *EvalExpr
	if i.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(i.Match)
		if err != nil {
			return nil, errW(err, "profile manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("project.items[%d].%s.imports[%d].match", itemIndex, scope, importIndex)),
				kv("value", i.Match),
			)
		}
	}

	return newProjectImportEntity(i.Link, i.Match, linkObj, matchObj), nil
}

// endregion
