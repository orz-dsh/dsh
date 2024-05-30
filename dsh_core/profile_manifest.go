package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
	"path/filepath"
	"regexp"
)

// region manifest

type ProfileManifest struct {
	Workspace    *ProfileManifestWorkspace
	Project      *ProfileManifestProject
	manifestPath string
	manifestType manifestMetadataType
}

func loadProfileManifest(path string) (*ProfileManifest, error) {
	manifest := &ProfileManifest{
		Workspace: NewProfileManifestWorkspace(nil, nil),
		Project:   NewProfileManifestProject(nil, nil, nil),
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

func MakeProfileManifest(workspace *ProfileManifestWorkspace, project *ProfileManifestProject) (*ProfileManifest, error) {
	if workspace == nil {
		workspace = NewProfileManifestWorkspace(nil, nil)
	}
	if project == nil {
		project = NewProfileManifestProject(nil, nil, nil)
	}
	manifest := &ProfileManifest{
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
		var matchExpr *vm.Program
		if item.Match != "" {
			matchExpr, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "profile manifest invalid",
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

		var matchExpr *vm.Program
		if registry.Match != "" {
			matchExpr, err = dsh_utils.CompileExpr(registry.Match)
			if err != nil {
				return errW(err, "profile manifest invalid",
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

		var matchObj *vm.Program
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
	Option *ProfileManifestProjectOption
	Script *ProfileManifestProjectScript
	Config *ProfileManifestProjectConfig
}

func NewProfileManifestProject(option *ProfileManifestProjectOption, script *ProfileManifestProjectScript, config *ProfileManifestProjectConfig) *ProfileManifestProject {
	if option == nil {
		option = &ProfileManifestProjectOption{}
	}
	if script == nil {
		script = &ProfileManifestProjectScript{}
	}
	if config == nil {
		config = &ProfileManifestProjectConfig{}
	}
	return &ProfileManifestProject{
		Option: option,
		Script: script,
		Config: config,
	}
}

func (p *ProfileManifestProject) init(manifest *ProfileManifest) (err error) {
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

type ProfileManifestProjectOption struct {
	Items    []*ProfileManifestProjectOptionItem
	entities projectOptionSpecifyEntitySet
}

type ProfileManifestProjectOptionItem struct {
	Name  string
	Value string
	Match string
}

func NewProfileManifestProjectOption(items map[string]string) *ProfileManifestProjectOption {
	var optionItems []*ProfileManifestProjectOptionItem
	for k, v := range items {
		optionItems = append(optionItems, &ProfileManifestProjectOptionItem{
			Name:  k,
			Value: v,
		})
	}
	return &ProfileManifestProjectOption{
		Items: optionItems,
	}
}

func (o *ProfileManifestProjectOption) init(manifest *ProfileManifest) (err error) {
	entities := projectOptionSpecifyEntitySet{}
	for i := 0; i < len(o.Items); i++ {
		item := o.Items[i]
		if item.Name == "" {
			return errN("profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("project.option.items[%d].name", i)),
			)
		}
		var matchObj *vm.Program
		if item.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "profile manifest invalid",
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

type ProfileManifestProjectScript struct {
	Sources        []*ProfileManifestProjectSource
	Imports        []*ProfileManifestProjectImport
	sourceEntities projectSourceEntitySet
	importEntities []*projectImportEntity
}

func NewProfileManifestProjectScript(sources []*ProfileManifestProjectSource, imports []*ProfileManifestProjectImport) *ProfileManifestProjectScript {
	return &ProfileManifestProjectScript{
		Sources: sources,
		Imports: imports,
	}
}

func (s *ProfileManifestProjectScript) init(manifest *ProfileManifest) (err error) {
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

type ProfileManifestProjectConfig struct {
	Sources        []*ProfileManifestProjectSource
	Imports        []*ProfileManifestProjectImport
	sourceEntities []*projectSourceEntity
	importEntities []*projectImportEntity
}

func NewProfileManifestProjectConfig(sources []*ProfileManifestProjectSource, imports []*ProfileManifestProjectImport) *ProfileManifestProjectConfig {
	return &ProfileManifestProjectConfig{
		Sources: sources,
		Imports: imports,
	}
}

func (c *ProfileManifestProjectConfig) init(manifest *ProfileManifest) (err error) {
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

func (s *ProfileManifestProjectSource) init(manifest *ProfileManifest, scope string, index int) (entity *projectSourceEntity, err error) {
	if s.Dir == "" {
		return nil, errN("profile manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.%s.sources[%d].dir", scope, index)),
		)
	}

	var matchObj *vm.Program
	if s.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(s.Match)
		if err != nil {
			return nil, errW(err, "profile manifest invalid",
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

func (i *ProfileManifestProjectImport) init(manifest *ProfileManifest, scope string, index int) (entity *projectImportEntity, err error) {
	if i.Link == "" {
		return nil, errN("profile manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("project.%s.imports[%d].link", scope, index)),
		)
	}
	linkObj, err := parseProjectLink(i.Link)
	if err != nil {
		return nil, errW(err, "profile manifest invalid",
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
			return nil, errW(err, "profile manifest invalid",
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
