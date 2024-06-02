package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"regexp"
)

// region profile

type ProfilePref struct {
	Option    *ProfileOptionPref
	Workspace *ProfileWorkspacePref
	Project   *ProfileProjectPref
	path      string
}

func loadProfilePref(path string) (pref *ProfilePref, error error) {
	pref = &ProfilePref{
		Option:    NewProfileOptionPref(nil),
		Workspace: NewProfileWorkspacePref(nil, nil),
		Project:   NewProfileProjectPref(nil),
	}

	metadata, err := dsh_utils.DeserializeFromFile(path, "", pref)
	if err != nil {
		return nil, errW(err, "load profile pref error",
			reason("deserialize error"),
			kv("path", path),
		)
	}
	pref.path = metadata.Path
	if err = pref.init(); err != nil {
		return nil, err
	}
	return pref, nil
}

func MakeProfilePref(option *ProfileOptionPref, workspace *ProfileWorkspacePref, project *ProfileProjectPref) (*ProfilePref, error) {
	if option == nil {
		option = NewProfileOptionPref(nil)
	}
	if workspace == nil {
		workspace = NewProfileWorkspacePref(nil, nil)
	}
	if project == nil {
		project = NewProfileProjectPref(nil)
	}
	pref := &ProfilePref{
		Option:    option,
		Workspace: workspace,
		Project:   project,
	}
	if err := pref.init(); err != nil {
		return nil, err
	}
	return pref, nil
}

func (p *ProfilePref) DescExtraKeyValues() KVS {
	return KVS{
		kv("path", p.path),
	}
}

func (p *ProfilePref) init() (err error) {
	if err = p.Option.init(p); err != nil {
		return err
	}
	if err = p.Workspace.init(p); err != nil {
		return err
	}
	if err = p.Project.init(p); err != nil {
		return err
	}
	return nil
}

// endregion

// region option

type ProfileOptionPref struct {
	Items   []*ProfileOptionItemPref
	schemas profileOptionSchemaSet
}

type ProfileOptionItemPref struct {
	Name  string
	Value string
	Match string
}

func NewProfileOptionPref(items []*ProfileOptionItemPref) *ProfileOptionPref {
	return &ProfileOptionPref{
		Items: items,
	}
}

func NewProfileOptionItemPref(name, value, match string) *ProfileOptionItemPref {
	return &ProfileOptionItemPref{
		Name:  name,
		Value: value,
		Match: match,
	}
}

func (p *ProfileOptionPref) init(pref *ProfilePref) error {
	schemas := profileOptionSchemaSet{}
	for i := 0; i < len(p.Items); i++ {
		if schema, err := p.Items[i].init(pref, i); err != nil {
			return err
		} else {
			schemas = append(schemas, schema)
		}
	}

	p.schemas = schemas
	return nil
}

func (p *ProfileOptionItemPref) init(pref *ProfilePref, itemIndex int) (schema *profileOptionSchema, err error) {
	if p.Name == "" {
		return nil, errN("profile pref invalid",
			reason("value empty"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("option.items[%d].name", itemIndex)),
		)
	}
	if checked := profileOptionNameCheckRegex.MatchString(p.Name); !checked {
		return nil, errN("profile pref invalid",
			reason("value invalid"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("option.items[%d].name", itemIndex)),
			kv("value", p.Name),
		)
	}

	var matchObj *EvalExpr
	if p.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(p.Match)
		if err != nil {
			return nil, errW(err, "profile pref invalid",
				reason("value invalid"),
				kv("path", pref.path),
				kv("field", fmt.Sprintf("option.items[%d].match", itemIndex)),
				kv("value", p.Match),
			)
		}
	}
	return newProfileOptionSchema(p.Name, p.Value, p.Match, matchObj), nil
}

// endregion

// region workspace

type ProfileWorkspacePref struct {
	Shell  *ProfileWorkspaceShellPref
	Import *ProfileWorkspaceImportPref
}

func NewProfileWorkspacePref(shell *ProfileWorkspaceShellPref, import_ *ProfileWorkspaceImportPref) *ProfileWorkspacePref {
	if shell == nil {
		shell = &ProfileWorkspaceShellPref{}
	}
	if import_ == nil {
		import_ = NewProfileWorkspaceImportPref(nil, nil)
	}
	return &ProfileWorkspacePref{
		Shell:  shell,
		Import: import_,
	}
}

func (p *ProfileWorkspacePref) init(pref *ProfilePref) (err error) {
	if err = p.Shell.init(pref); err != nil {
		return err
	}
	if err = p.Import.init(pref); err != nil {
		return err
	}
	return nil
}

// endregion

// region workspace shell

type ProfileWorkspaceShellPref struct {
	Items   []*ProfileWorkspaceShellItemPref
	schemas workspaceShellSettingSet
}

type ProfileWorkspaceShellItemPref struct {
	Name  string
	Path  string
	Exts  []string
	Args  []string
	Match string
}

func NewProfileWorkspaceShellPref(items []*ProfileWorkspaceShellItemPref) *ProfileWorkspaceShellPref {
	return &ProfileWorkspaceShellPref{
		Items: items,
	}
}

func NewProfileWorkspaceShellItemPref(name, path string, exts, args []string, match string) *ProfileWorkspaceShellItemPref {
	return &ProfileWorkspaceShellItemPref{
		Name:  name,
		Path:  path,
		Exts:  exts,
		Args:  args,
		Match: match,
	}
}

func (p *ProfileWorkspaceShellPref) init(pref *ProfilePref) (err error) {
	schemas := workspaceShellSettingSet{}
	for i := 0; i < len(p.Items); i++ {
		item := p.Items[i]
		if item.Name == "" {
			return errN("profile pref invalid",
				reason("value empty"),
				kv("path", pref.path),
				kv("field", fmt.Sprintf("workspace.shell.items[%d].name", i)),
			)
		}
		if item.Path != "" && !dsh_utils.IsFileExists(item.Path) {
			return errN("profile pref invalid",
				reason("value invalid"),
				kv("path", pref.path),
				kv("field", fmt.Sprintf("workspace.shell.items[%d].path", i)),
				kv("value", item.Path),
			)
		}
		for j := 0; j < len(item.Exts); j++ {
			if item.Exts[j] == "" {
				return errN("profile pref invalid",
					reason("value empty"),
					kv("path", pref.path),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].exts[%d]", i, j)),
				)
			}
		}
		for j := 0; j < len(item.Args); j++ {
			if item.Args[j] == "" {
				return errN("profile pref invalid",
					reason("value empty"),
					kv("path", pref.path),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].args[%d]", i, j)),
				)
			}
		}
		var matchObj *EvalExpr
		if item.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "profile pref invalid",
					reason("value invalid"),
					kv("path", pref.path),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].match", i)),
					kv("value", item.Match),
				)
			}
		}
		schemas[item.Name] = append(schemas[item.Name], newWorkspaceShellSetting(item.Name, item.Path, item.Exts, item.Args, item.Match, matchObj))
	}

	p.schemas = schemas
	return nil
}

// endregion

// region workspace import

type ProfileWorkspaceImportPref struct {
	Registry        *ProfileImportRegistryPref
	Redirect        *ProfileImportRedirectPref
	registrySchemas workspaceImportRegistrySettingSet
	redirectSchemas workspaceImportRedirectSettingSet
}

func NewProfileWorkspaceImportPref(registry *ProfileImportRegistryPref, redirect *ProfileImportRedirectPref) *ProfileWorkspaceImportPref {
	if registry == nil {
		registry = NewProfileImportRegistryPref(nil)
	}
	if redirect == nil {
		redirect = NewProfileImportRedirectPref(nil)
	}
	return &ProfileWorkspaceImportPref{
		Registry: registry,
		Redirect: redirect,
	}
}

func (p *ProfileWorkspaceImportPref) init(pref *ProfilePref) (err error) {
	var registrySchemas workspaceImportRegistrySettingSet
	if registrySchemas, err = p.Registry.init(pref); err != nil {
		return err
	}

	var redirectSchemas workspaceImportRedirectSettingSet
	if redirectSchemas, err = p.Redirect.init(pref); err != nil {
		return err
	}

	p.registrySchemas = registrySchemas
	p.redirectSchemas = redirectSchemas
	return nil
}

// endregion

// region workspace import registry

type ProfileImportRegistryPref struct {
	Items []*ProfileImportRegistryItemPref
}

type ProfileImportRegistryItemPref struct {
	Name  string
	Link  string
	Match string
}

func NewProfileImportRegistryPref(items []*ProfileImportRegistryItemPref) *ProfileImportRegistryPref {
	return &ProfileImportRegistryPref{
		Items: items,
	}
}

func NewProfileImportRegistryItemPref(name, link, match string) *ProfileImportRegistryItemPref {
	return &ProfileImportRegistryItemPref{
		Name:  name,
		Link:  link,
		Match: match,
	}
}

func (p *ProfileImportRegistryPref) init(pref *ProfilePref) (workspaceImportRegistrySettingSet, error) {
	schemas := workspaceImportRegistrySettingSet{}
	for i := 0; i < len(p.Items); i++ {
		item := p.Items[i]
		if schema, err := item.init(pref, i); err != nil {
			return nil, err
		} else {
			schemas[item.Name] = append(schemas[item.Name], schema)
		}
	}
	return schemas, nil
}

func (p *ProfileImportRegistryItemPref) init(pref *ProfilePref, itemIndex int) (schema *workspaceImportRegistrySetting, err error) {
	if p.Name == "" {
		return nil, errN("profile pref invalid",
			reason("value empty"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("workspace.import.registry.items[%d].name", itemIndex)),
		)
	}

	if p.Link == "" {
		return nil, errN("profile pref invalid",
			reason("value empty"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("workspace.import.registry.items[%d].link", itemIndex)),
		)
	}
	// TODO: check link template

	var matchObj *EvalExpr
	if p.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(p.Match)
		if err != nil {
			return nil, errW(err, "profile pref invalid",
				reason("value invalid"),
				kv("path", pref.path),
				kv("field", fmt.Sprintf("workspace.import.registry.items[%d].match", itemIndex)),
				kv("value", p.Match),
			)
		}
	}

	return newWorkspaceImportRegistrySetting(p.Name, p.Link, p.Match, matchObj), nil
}

// endregion

// region workspace import redirect

type ProfileImportRedirectPref struct {
	Items []*ProfileImportRedirectItemPref
}

type ProfileImportRedirectItemPref struct {
	Regex string
	Link  string
	Match string
}

func NewProfileImportRedirectPref(items []*ProfileImportRedirectItemPref) *ProfileImportRedirectPref {
	return &ProfileImportRedirectPref{
		Items: items,
	}
}

func NewProfileImportRedirectItemPref(regex, link, match string) *ProfileImportRedirectItemPref {
	return &ProfileImportRedirectItemPref{
		Regex: regex,
		Link:  link,
		Match: match,
	}
}

func (p *ProfileImportRedirectPref) init(pref *ProfilePref) (workspaceImportRedirectSettingSet, error) {
	schemas := workspaceImportRedirectSettingSet{}
	for i := 0; i < len(p.Items); i++ {
		item := p.Items[i]
		if schema, err := item.init(pref, i); err != nil {
			return nil, err
		} else {
			schemas = append(schemas, schema)
		}
	}
	return schemas, nil
}

func (p *ProfileImportRedirectItemPref) init(pref *ProfilePref, itemIndex int) (schema *workspaceImportRedirectSetting, err error) {
	if p.Regex == "" {
		return nil, errN("profile pref invalid",
			reason("value empty"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("workspace.import.redirect.items[%d].regex", itemIndex)),
		)
	}
	regexObj, err := regexp.Compile(p.Regex)
	if err != nil {
		return nil, errW(err, "profile pref invalid",
			reason("value invalid"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("workspace.import.redirect.items[%d].regex", itemIndex)),
			kv("value", p.Regex),
		)
	}

	if p.Link == "" {
		return nil, errN("profile pref invalid",
			reason("value empty"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("workspace.import.redirect.items[%d].link", itemIndex)),
		)
	}
	// TODO: check link template

	var matchObj *EvalExpr
	if p.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(p.Match)
		if err != nil {
			return nil, errW(err, "profile pref invalid",
				reason("value invalid"),
				kv("path", pref.path),
				kv("field", fmt.Sprintf("workspace.import.redirect.items[%d].match", itemIndex)),
				kv("value", p.Match),
			)
		}
	}

	return newWorkspaceImportRedirectSetting(p.Regex, p.Link, p.Match, regexObj, matchObj), nil
}

// endregion

// region project

type ProfileProjectPref struct {
	Items   []*ProfileProjectItemPref
	schemas profileProjectSchemaSet
}

type ProfileProjectItemPref struct {
	Name   string
	Path   string
	Match  string
	Script *ProfileProjectScriptPref
	Config *ProfileProjectConfigPref
}

func NewProfileProjectPref(items []*ProfileProjectItemPref) *ProfileProjectPref {
	return &ProfileProjectPref{
		Items: items,
	}
}

func NewProfileProjectItemPref(script *ProfileProjectScriptPref, config *ProfileProjectConfigPref) *ProfileProjectItemPref {
	if script == nil {
		script = &ProfileProjectScriptPref{}
	}
	if config == nil {
		config = &ProfileProjectConfigPref{}
	}
	return &ProfileProjectItemPref{
		Script: script,
		Config: config,
	}
}

func (p *ProfileProjectPref) init(pref *ProfilePref) error {
	schemas := profileProjectSchemaSet{}
	for i := 0; i < len(p.Items); i++ {
		if schema, err := p.Items[i].init(pref, i); err != nil {
			return err
		} else {
			schemas = append(schemas, schema)
		}
	}

	p.schemas = schemas
	return nil
}

func (p *ProfileProjectItemPref) init(pref *ProfilePref, itemIndex int) (schema *profileProjectSchema, err error) {
	if p.Name == "" {
		return nil, errN("profile pref invalid",
			reason("value empty"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("project.items[%d].name", itemIndex)),
		)
	}
	if checked := projectNameCheckRegex.MatchString(p.Name); !checked {
		return nil, errN("profile pref invalid",
			reason("value invalid"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("project.items[%d].name", itemIndex)),
			kv("value", p.Name),
		)
	}

	if p.Path == "" {
		return nil, errN("profile pref invalid",
			reason("value empty"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("project.items[%d].path", itemIndex)),
		)
	}

	var scriptSourceSchemas projectSchemaSourceSet
	var scriptImportSchemas projectSchemaImportSet
	if p.Script != nil {
		if scriptSourceSchemas, scriptImportSchemas, err = p.Script.init(pref, itemIndex); err != nil {
			return nil, err
		}
	}

	var configSourceSchemas projectSchemaSourceSet
	var configImportSchemas projectSchemaImportSet
	if p.Config != nil {
		if configSourceSchemas, configImportSchemas, err = p.Config.init(pref, itemIndex); err != nil {
			return nil, err
		}
	}

	var matchObj *EvalExpr
	if p.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(p.Match)
		if err != nil {
			return nil, errW(err, "profile pref invalid",
				reason("value invalid"),
				kv("path", pref.path),
				kv("field", fmt.Sprintf("project.items[%d].match", itemIndex)),
				kv("value", p.Match),
			)
		}
	}

	schema = newProfileProjectSchema(p.Name, p.Path, p.Match, scriptSourceSchemas, scriptImportSchemas, configSourceSchemas, configImportSchemas, matchObj)
	return schema, nil
}

// endregion

// region project script

type ProfileProjectScriptPref struct {
	Sources []*ProfileProjectSourcePref
	Imports []*ProfileProjectImportPref
}

func NewProfileProjectScriptPref(sources []*ProfileProjectSourcePref, imports []*ProfileProjectImportPref) *ProfileProjectScriptPref {
	return &ProfileProjectScriptPref{
		Sources: sources,
		Imports: imports,
	}
}

func (p *ProfileProjectScriptPref) init(pref *ProfilePref, itemIndex int) (projectSchemaSourceSet, projectSchemaImportSet, error) {
	sourceSchemas := projectSchemaSourceSet{}
	for sourceIndex := 0; sourceIndex < len(p.Sources); sourceIndex++ {
		src := p.Sources[sourceIndex]
		if schemas, err := src.init(pref, "script", itemIndex, sourceIndex); err != nil {
			return nil, nil, err
		} else {
			sourceSchemas = append(sourceSchemas, schemas)
		}
	}

	importSchemas := projectSchemaImportSet{}
	for importIndex := 0; importIndex < len(p.Imports); importIndex++ {
		imp := p.Imports[importIndex]
		if schema, err := imp.init(pref, "script", itemIndex, importIndex); err != nil {
			return nil, nil, err
		} else {
			importSchemas = append(importSchemas, schema)
		}
	}

	return sourceSchemas, importSchemas, nil
}

// endregion

// region project config

type ProfileProjectConfigPref struct {
	Sources []*ProfileProjectSourcePref
	Imports []*ProfileProjectImportPref
}

func NewProfileProjectConfigPref(sources []*ProfileProjectSourcePref, imports []*ProfileProjectImportPref) *ProfileProjectConfigPref {
	return &ProfileProjectConfigPref{
		Sources: sources,
		Imports: imports,
	}
}

func (p *ProfileProjectConfigPref) init(pref *ProfilePref, itemIndex int) (projectSchemaSourceSet, projectSchemaImportSet, error) {
	sourceSchemas := projectSchemaSourceSet{}
	for sourceIndex := 0; sourceIndex < len(p.Sources); sourceIndex++ {
		src := p.Sources[sourceIndex]
		if schema, err := src.init(pref, "config", itemIndex, sourceIndex); err != nil {
			return nil, nil, err
		} else {
			sourceSchemas = append(sourceSchemas, schema)
		}
	}

	importSchemas := projectSchemaImportSet{}
	for importIndex := 0; importIndex < len(p.Imports); importIndex++ {
		imp := p.Imports[importIndex]
		if schema, err := imp.init(pref, "config", itemIndex, importIndex); err != nil {
			return nil, nil, err
		} else {
			importSchemas = append(importSchemas, schema)
		}
	}

	return sourceSchemas, importSchemas, nil
}

// endregion

// region project source

type ProfileProjectSourcePref struct {
	Dir   string
	Files []string
	Match string
}

func NewProfileProjectSourcePref(dir string, files []string, match string) *ProfileProjectSourcePref {
	return &ProfileProjectSourcePref{
		Dir:   dir,
		Files: files,
		Match: match,
	}
}

func (p *ProfileProjectSourcePref) init(pref *ProfilePref, scope string, itemIndex, sourceIndex int) (schema *projectSchemaSource, err error) {
	if p.Dir == "" {
		return nil, errN("profile pref invalid",
			reason("value empty"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("project.items[%d].%s.sources[%d].dir", itemIndex, scope, sourceIndex)),
		)
	}

	var matchObj *EvalExpr
	if p.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(p.Match)
		if err != nil {
			return nil, errW(err, "profile pref invalid",
				reason("value invalid"),
				kv("path", pref.path),
				kv("field", fmt.Sprintf("project.items[%d].%s.sources[%d].match", itemIndex, scope, sourceIndex)),
				kv("value", p.Match),
			)
		}
	}

	return newProjectSchemaSource(p.Dir, p.Files, p.Match, matchObj), nil
}

// endregion

// region project import

type ProfileProjectImportPref struct {
	Link  string
	Match string
}

func NewProfileProjectImportPref(link string, match string) *ProfileProjectImportPref {
	return &ProfileProjectImportPref{
		Link:  link,
		Match: match,
	}
}

func (p *ProfileProjectImportPref) init(pref *ProfilePref, scope string, itemIndex, importIndex int) (schema *projectSchemaImport, err error) {
	if p.Link == "" {
		return nil, errN("profile pref invalid",
			reason("value empty"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("project.items[%d].%s.imports[%d].link", itemIndex, scope, importIndex)),
		)
	}
	linkObj, err := parseProjectLink(p.Link)
	if err != nil {
		return nil, errW(err, "profile pref invalid",
			reason("value invalid"),
			kv("path", pref.path),
			kv("field", fmt.Sprintf("project.items[%d].%s.imports[%d].link", itemIndex, scope, importIndex)),
			kv("value", p.Link),
		)
	}

	var matchObj *EvalExpr
	if p.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(p.Match)
		if err != nil {
			return nil, errW(err, "profile pref invalid",
				reason("value invalid"),
				kv("path", pref.path),
				kv("field", fmt.Sprintf("project.items[%d].%s.imports[%d].match", itemIndex, scope, importIndex)),
				kv("value", p.Match),
			)
		}
	}

	return newProjectSchemaImport(p.Link, p.Match, linkObj, matchObj), nil
}

// endregion
