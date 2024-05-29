package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
	"path/filepath"
	"regexp"
)

// region manifest

type AppProfileManifest struct {
	Workspace    *AppProfileManifestWorkspace
	Project      *AppProfileManifestProject
	manifestPath string
	manifestType manifestMetadataType
}

func loadAppProfileManifest(path string) (*AppProfileManifest, error) {
	manifest := &AppProfileManifest{
		Workspace: NewAppProfileManifestWorkspace(nil, nil),
		Project:   NewAppProfileManifestProject(nil, nil, nil),
	}

	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, errW(err, "load app profile manifest error",
				reason("get abs-path error"),
				kv("path", path),
			)
		}
		path = absPath
	}

	if path != "" {
		metadata, err := loadManifestFromFile(path, "", manifest)
		if err != nil {
			return nil, errW(err, "load app profile manifest error",
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

func MakeAppProfileManifest(workspace *AppProfileManifestWorkspace, project *AppProfileManifestProject) (*AppProfileManifest, error) {
	if workspace == nil {
		workspace = NewAppProfileManifestWorkspace(nil, nil)
	}
	if project == nil {
		project = NewAppProfileManifestProject(nil, nil, nil)
	}
	manifest := &AppProfileManifest{
		Workspace: workspace,
		Project:   project,
	}

	if err := manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *AppProfileManifest) DescExtraKeyValues() KVS {
	return KVS{
		kv("manifestPath", m.manifestPath),
		kv("manifestType", m.manifestType),
	}
}

func (m *AppProfileManifest) init() (err error) {
	if err = m.Workspace.init(m); err != nil {
		return err
	}
	if err = m.Project.init(m); err != nil {
		return err
	}

	return nil
}

// endregion

// region workspace

type AppProfileManifestWorkspace struct {
	Shell  *AppProfileManifestWorkspaceShell
	Import *AppProfileManifestWorkspaceImport
}

func NewAppProfileManifestWorkspace(shell *AppProfileManifestWorkspaceShell, imp *AppProfileManifestWorkspaceImport) *AppProfileManifestWorkspace {
	if shell == nil {
		shell = &AppProfileManifestWorkspaceShell{}
	}
	if imp == nil {
		imp = &AppProfileManifestWorkspaceImport{}
	}
	return &AppProfileManifestWorkspace{
		Shell:  shell,
		Import: imp,
	}
}

func (w *AppProfileManifestWorkspace) init(manifest *AppProfileManifest) (err error) {
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

type AppProfileManifestWorkspaceShell struct {
	Items    []*AppProfileManifestWorkspaceShellItem
	entities workspaceShellEntitySet
}

type AppProfileManifestWorkspaceShellItem struct {
	Name  string
	Path  string
	Exts  []string
	Args  []string
	Match string
}

func (s *AppProfileManifestWorkspaceShell) init(manifest *AppProfileManifest) (err error) {
	entities := workspaceShellEntitySet{}
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if item.Name == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.shell.items[%d].name", i)),
			)
		}
		if item.Path != "" && !dsh_utils.IsFileExists(item.Path) {
			return errN("app profile manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.shell.items[%d].path", i)),
				kv("value", item.Path),
			)
		}
		for j := 0; j < len(item.Exts); j++ {
			if item.Exts[j] == "" {
				return errN("app profile manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].exts[%d]", i, j)),
				)
			}
		}
		for j := 0; j < len(item.Args); j++ {
			if item.Args[j] == "" {
				return errN("app profile manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].args[%d]", i, j)),
				)
			}
		}
		var matchExpr *vm.Program
		if item.Match != "" {
			matchExpr, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "app profile manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].match", i)),
					kv("value", item.Match),
				)
			}
		}
		entities[item.Name] = append(entities[item.Name], newWorkspaceShellEntity(item.Name, item.Path, item.Exts, item.Args, item.Match, matchExpr))
	}

	s.entities = entities
	return nil
}

// endregion

// region workspace import

type AppProfileManifestWorkspaceImport struct {
	Registries       []*AppProfileManifestImportRegistry
	Redirects        []*AppProfileManifestImportRedirect
	registryEntities workspaceImportRegistryEntitySet
	redirectEntities workspaceImportRedirectEntitySet
}

type AppProfileManifestImportRegistry struct {
	Name  string
	Link  string
	Match string
}

type AppProfileManifestImportRedirect struct {
	Regex string
	Link  string
	Match string
}

func (imp *AppProfileManifestWorkspaceImport) init(manifest *AppProfileManifest) (err error) {
	registryEntities := workspaceImportRegistryEntitySet{}
	for i := 0; i < len(imp.Registries); i++ {
		registry := imp.Registries[i]
		if registry.Name == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.registries[%d].name", i)),
			)
		}

		if registry.Link == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.registries[%d].link", i)),
			)
		}
		// TODO: check link template

		var matchExpr *vm.Program
		if registry.Match != "" {
			matchExpr, err = dsh_utils.CompileExpr(registry.Match)
			if err != nil {
				return errW(err, "app profile manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.import.registries[%d].match", i)),
					kv("value", registry.Match),
				)
			}
		}
		registryEntities[registry.Name] = append(registryEntities[registry.Name], newWorkspaceImportRegistryEntity(registry.Name, registry.Link, registry.Match, matchExpr))
	}

	redirectEntities := workspaceImportRedirectEntitySet{}
	for i := 0; i < len(imp.Redirects); i++ {
		redirect := imp.Redirects[i]
		if redirect.Regex == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.redirects[%d].regex", i)),
			)
		}
		regexObj, err := regexp.Compile(redirect.Regex)
		if err != nil {
			return errW(err, "app profile manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.redirects[%d].regex", i)),
				kv("value", redirect.Regex),
			)
		}

		if redirect.Link == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("workspace.import.redirects[%d].link", i)),
			)
		}
		// TODO: check link template

		var matchObj *vm.Program
		if redirect.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(redirect.Match)
			if err != nil {
				return errW(err, "app profile manifest invalid",
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

type AppProfileManifestProject struct {
	Option *AppProfileManifestProjectOption
	Script *AppProfileManifestProjectScript
	Config *AppProfileManifestProjectConfig
}

func NewAppProfileManifestProject(option *AppProfileManifestProjectOption, script *AppProfileManifestProjectScript, config *AppProfileManifestProjectConfig) *AppProfileManifestProject {
	if option == nil {
		option = &AppProfileManifestProjectOption{}
	}
	if script == nil {
		script = &AppProfileManifestProjectScript{}
	}
	if config == nil {
		config = &AppProfileManifestProjectConfig{}
	}
	return &AppProfileManifestProject{
		Option: option,
		Script: script,
		Config: config,
	}
}

func (p *AppProfileManifestProject) init(manifest *AppProfileManifest) (err error) {
	if err = p.Option.init(manifest); err != nil {
		return err
	}
	if err = p.Script.init(manifest); err != nil {
		return err
	}
	if err = p.Config.init(manifest); err != nil {
		return err
	}
	return nil
}

// endregion

// region project option

type AppProfileManifestProjectOption struct {
	Items    []*AppProfileManifestProjectOptionItem
	entities projectOptionSpecifyEntitySet
}

type AppProfileManifestProjectOptionItem struct {
	Name  string
	Value string
	Match string
}

func NewAppProfileManifestProjectOption(items map[string]string) *AppProfileManifestProjectOption {
	var optionItems []*AppProfileManifestProjectOptionItem
	for k, v := range items {
		optionItems = append(optionItems, &AppProfileManifestProjectOptionItem{
			Name:  k,
			Value: v,
		})
	}
	return &AppProfileManifestProjectOption{
		Items: optionItems,
	}
}

func (o *AppProfileManifestProjectOption) init(manifest *AppProfileManifest) (err error) {
	entities := projectOptionSpecifyEntitySet{}
	for i := 0; i < len(o.Items); i++ {
		item := o.Items[i]
		if item.Name == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("project.option.items[%d].name", i)),
			)
		}
		var matchObj *vm.Program
		if item.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "app profile manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("project.option.items[%d].match", i)),
					kv("value", item.Match),
				)
			}
		}
		entities = append(entities, newProjectOptionSpecifyEntity(item.Name, item.Value, item.Match, matchObj))
	}

	o.entities = entities
	return nil
}

// endregion

// region project script

type AppProfileManifestProjectScript struct {
	Sources        []*AppProfileManifestProjectSource
	Imports        []*AppProfileManifestProjectImport
	sourceEntities projectSourceEntitySet
	importEntities []*projectImportEntity
}

func NewAppProfileManifestProjectScript(sources []*AppProfileManifestProjectSource, imports []*AppProfileManifestProjectImport) *AppProfileManifestProjectScript {
	return &AppProfileManifestProjectScript{
		Sources: sources,
		Imports: imports,
	}
}

func (s *AppProfileManifestProjectScript) init(manifest *AppProfileManifest) (err error) {
	sourceEntities := projectSourceEntitySet{}
	for i := 0; i < len(s.Sources); i++ {
		src := s.Sources[i]
		if entity, err := src.init(manifest, "script", i); err != nil {
			return err
		} else {
			sourceEntities = append(sourceEntities, entity)
		}
	}

	importEntities := projectImportEntitySet{}
	for i := 0; i < len(s.Imports); i++ {
		imp := s.Imports[i]
		if entity, err := imp.init(manifest, "script", i); err != nil {
			return err
		} else {
			importEntities = append(importEntities, entity)
		}
	}

	s.sourceEntities = sourceEntities
	s.importEntities = importEntities
	return nil
}

// endregion

// region project config

type AppProfileManifestProjectConfig struct {
	Sources        []*AppProfileManifestProjectSource
	Imports        []*AppProfileManifestProjectImport
	sourceEntities []*projectSourceEntity
	importEntities []*projectImportEntity
}

func NewAppProfileManifestProjectConfig(sources []*AppProfileManifestProjectSource, imports []*AppProfileManifestProjectImport) *AppProfileManifestProjectConfig {
	return &AppProfileManifestProjectConfig{
		Sources: sources,
		Imports: imports,
	}
}

func (c *AppProfileManifestProjectConfig) init(manifest *AppProfileManifest) (err error) {
	sourceEntities := projectSourceEntitySet{}
	for i := 0; i < len(c.Sources); i++ {
		src := c.Sources[i]
		if entity, err := src.init(manifest, "config", i); err != nil {
			return err
		} else {
			sourceEntities = append(sourceEntities, entity)
		}
	}

	var importEntities []*projectImportEntity
	for i := 0; i < len(c.Imports); i++ {
		imp := c.Imports[i]
		if entity, err := imp.init(manifest, "config", i); err != nil {
			return err
		} else {
			importEntities = append(importEntities, entity)
		}
	}

	c.sourceEntities = sourceEntities
	c.importEntities = importEntities
	return nil
}

// endregion

// region project source

type AppProfileManifestProjectSource struct {
	Dir   string
	Files []string
	Match string
}

func NewAppProfileManifestProjectSource(dir string, files []string, match string) *AppProfileManifestProjectSource {
	return &AppProfileManifestProjectSource{
		Dir:   dir,
		Files: files,
		Match: match,
	}
}

func (s *AppProfileManifestProjectSource) init(manifest *AppProfileManifest, scope string, index int) (entity *projectSourceEntity, err error) {
	if s.Dir == "" {
		return nil, errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.%s.sources[%d].dir", scope, index)),
		)
	}

	var matchObj *vm.Program
	if s.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(s.Match)
		if err != nil {
			return nil, errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("project.%s.sources[%d].match", scope, index)),
				kv("value", s.Match),
			)
		}
	}

	return newProjectSourceEntity(s.Dir, s.Files, s.Match, matchObj), nil
}

// endregion

// region project import

type AppProfileManifestProjectImport struct {
	Link  string
	Match string
}

func NewAppProfileManifestProjectImport(link string, match string) *AppProfileManifestProjectImport {
	return &AppProfileManifestProjectImport{
		Link:  link,
		Match: match,
	}
}

func (i *AppProfileManifestProjectImport) init(manifest *AppProfileManifest, scope string, index int) (entity *projectImportEntity, err error) {
	if i.Link == "" {
		return nil, errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.%s.imports[%d].link", scope, index)),
		)
	}
	linkObj, err := ParseProjectLink(i.Link)
	if err != nil {
		return nil, errW(err, "project manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.%s.imports[%d].link", scope, index)),
			kv("value", i.Link),
		)
	}

	var matchObj *vm.Program
	if i.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(i.Match)
		if err != nil {
			return nil, errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("project.%s.imports[%d].match", scope, index)),
				kv("value", i.Match),
			)
		}
	}

	return newProjectImportEntity(i.Link, i.Match, linkObj, matchObj), nil
}

// endregion
