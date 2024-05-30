package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
	"regexp"
)

// region manifest

var projectNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*$")

type ProjectManifest struct {
	Name         string
	Runtime      *ProjectManifestRuntime
	Option       *ProjectManifestOption
	Script       *ProjectManifestScript
	Config       *ProjectManifestConfig
	manifestPath string
	manifestType manifestMetadataType
	projectName  string
	projectPath  string
}

func loadProjectManifest(projectPath string) (manifest *ProjectManifest, err error) {
	manifest = &ProjectManifest{
		Runtime: &ProjectManifestRuntime{},
		Option:  &ProjectManifestOption{},
		Script:  &ProjectManifestScript{},
		Config:  &ProjectManifestConfig{},
	}
	metadata, err := loadManifestFromDir(projectPath, []string{"project"}, manifest, true)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("load manifest from dir error"),
			kv("projectPath", projectPath),
		)
	}
	manifest.manifestPath = metadata.ManifestPath
	manifest.manifestType = metadata.ManifestType
	manifest.projectPath = projectPath
	if err = manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func MakeProjectManifest(name string, runtime *ProjectManifestRuntime, option *ProjectManifestOption, script *ProjectManifestScript, config *ProjectManifestConfig) (manifest *ProjectManifest, err error) {
	if runtime == nil {
		runtime = &ProjectManifestRuntime{}
	}
	if option == nil {
		option = &ProjectManifestOption{}
	}
	if script == nil {
		script = &ProjectManifestScript{}
	}
	if config == nil {
		config = &ProjectManifestConfig{}
	}
	manifest = &ProjectManifest{
		Name:    name,
		Runtime: runtime,
		Option:  option,
		Script:  script,
		Config:  config,
	}
	if err = manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *ProjectManifest) DescExtraKeyValues() KVS {
	return KVS{
		kv("projectName", m.projectName),
		kv("projectPath", m.projectPath),
		kv("manifestPath", m.manifestPath),
		kv("manifestType", m.manifestType),
	}
}

func (m *ProjectManifest) init() (err error) {
	if m.Name == "" {
		return errN("project manifest invalid",
			reason("name empty"),
			kv("path", m.manifestPath),
			kv("field", "name"),
		)
	}
	if checked := projectNameCheckRegex.MatchString(m.Name); !checked {
		return errN("project manifest invalid",
			reason("value invalid"),
			kv("path", m.manifestPath),
			kv("field", "name"),
			kv("value", m.Name),
		)
	}
	m.projectName = m.Name
	if err = m.Runtime.init(m); err != nil {
		return err
	}
	if err = m.Option.init(m); err != nil {
		return err
	}
	if err = m.Script.init(m); err != nil {
		return err
	}
	if err = m.Config.init(m); err != nil {
		return err
	}
	return nil
}

// endregion

// region runtime

type ProjectManifestRuntime struct {
	MinVersion dsh_utils.Version `yaml:"minVersion" toml:"minVersion" json:"minVersion"`
	MaxVersion dsh_utils.Version `yaml:"maxVersion" toml:"maxVersion" json:"maxVersion"`
}

func NewProjectManifestRuntime(minVersion dsh_utils.Version, maxVersion dsh_utils.Version) *ProjectManifestRuntime {
	return &ProjectManifestRuntime{
		MinVersion: minVersion,
		MaxVersion: maxVersion,
	}
}

func (r *ProjectManifestRuntime) init(manifest *ProjectManifest) (err error) {
	if err = dsh_utils.CheckRuntimeVersion(r.MinVersion, r.MaxVersion); err != nil {
		return errW(err, "project manifest invalid",
			reason("runtime incompatible"),
			kv("path", manifest.manifestPath),
			kv("field", "runtime"),
			kv("minVersion", r.MinVersion),
			kv("maxVersion", r.MaxVersion),
			kv("runtimeVersion", dsh_utils.GetRuntimeVersion()),
		)
	}
	return nil
}

// endregion

// region option

type ProjectManifestOption struct {
	Items           []*ProjectManifestOptionItem
	Verifies        []string
	declareEntities projectOptionDeclareEntitySet
	verifyEntities  projectOptionVerifyEntitySet
}

type ProjectManifestOptionItem struct {
	Name     string
	Type     projectOptionValueType
	Choices  []string
	Default  *string
	Optional bool
	Assigns  []*ProjectManifestOptionItemAssign
}

type ProjectManifestOptionItemAssign struct {
	Project string
	Option  string
	Mapping string
}

func NewProjectManifestOption(items []*ProjectManifestOptionItem, verifies []string) *ProjectManifestOption {
	return &ProjectManifestOption{
		Items:    items,
		Verifies: verifies,
	}
}

func NewProjectManifestOptionItem(name string, valueType projectOptionValueType, choices []string, defaultValue *string, optional bool, assigns []*ProjectManifestOptionItemAssign) *ProjectManifestOptionItem {
	return &ProjectManifestOptionItem{
		Name:     name,
		Type:     valueType,
		Choices:  choices,
		Default:  defaultValue,
		Optional: optional,
		Assigns:  assigns,
	}
}

func NewProjectManifestOptionItemAssign(project string, option string, mapping string) *ProjectManifestOptionItemAssign {
	return &ProjectManifestOptionItemAssign{
		Project: project,
		Option:  option,
		Mapping: mapping,
	}
}

func (o *ProjectManifestOption) init(manifest *ProjectManifest) error {
	declareEntities := projectOptionDeclareEntitySet{}
	optionNamesDict := map[string]bool{}
	assignTargetsDict := map[string]bool{}
	for i := 0; i < len(o.Items); i++ {
		if declareEntity, err := o.Items[i].init(manifest, optionNamesDict, assignTargetsDict, i); err != nil {
			return err
		} else {
			declareEntities = append(declareEntities, declareEntity)
		}
	}

	verifyEntities := projectOptionVerifyEntitySet{}
	for i := 0; i < len(o.Verifies); i++ {
		expr := o.Verifies[i]
		if expr == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
			)
		}
		exprObj, err := dsh_utils.CompileExpr(expr)
		if err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
				kv("value", expr),
			)
		}
		verifyEntities = append(verifyEntities, newProjectOptionVerifyEntity(expr, exprObj))
	}

	o.declareEntities = declareEntities
	o.verifyEntities = verifyEntities
	return nil
}

func (i *ProjectManifestOptionItem) init(manifest *ProjectManifest, itemNamesDict, assignTargetsDict map[string]bool, itemIndex int) (entity *projectOptionDeclareEntity, err error) {
	if i.Name == "" {
		return nil, errN("project manifest invalid",
			reason("name empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].name", itemIndex)),
		)
	}
	if checked := projectNameCheckRegex.MatchString(i.Name); !checked {
		return nil, errN("project manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].name", itemIndex)),
			kv("value", i.Name),
		)
	}
	if _, exist := itemNamesDict[i.Name]; exist {
		return nil, errN("project manifest invalid",
			reason("name duplicated"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].name", itemIndex)),
			kv("value", i.Name),
		)
	}
	valueType := i.Type
	if valueType == "" {
		valueType = projectOptionValueTypeString
	}
	switch valueType {
	case projectOptionValueTypeString:
	case projectOptionValueTypeBool:
	case projectOptionValueTypeInteger:
	case projectOptionValueTypeDecimal:
	default:
		return nil, errN("project manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].type", itemIndex)),
			kv("value", i.Type),
		)
	}
	entity = newProjectOptionDeclareEntity(i.Name, valueType, i.Choices, i.Optional)
	if err = entity.setDefaultValue(i.Default); err != nil {
		return nil, errW(err, "project manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].default", itemIndex)),
			kv("value", *i.Default),
		)
	}

	for assignIndex := 0; assignIndex < len(i.Assigns); assignIndex++ {
		assign := i.Assigns[assignIndex]
		if assignEntity, err := assign.init(manifest, assignTargetsDict, itemIndex, assignIndex); err != nil {
			return nil, err
		} else {
			entity.addAssign(assignEntity)
		}
	}

	itemNamesDict[i.Name] = true
	return entity, nil
}

func (a *ProjectManifestOptionItemAssign) init(manifest *ProjectManifest, targetsDict map[string]bool, itemIndex int, assignIndex int) (entity *projectOptionAssignEntity, err error) {
	if a.Project == "" {
		return nil, errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", itemIndex, assignIndex)),
		)
	}
	if a.Project == manifest.Name {
		return nil, errN("project manifest invalid",
			reason("can not assign to self project option"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", itemIndex, assignIndex)),
		)
	}
	if a.Option == "" {
		return nil, errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].assigns[%d].option", itemIndex, assignIndex)),
		)
	}
	assignTarget := a.Project + "." + a.Option
	if _, exists := targetsDict[assignTarget]; exists {
		return nil, errN("project manifest invalid",
			reason("option assign target duplicated"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].assigns[%d]", itemIndex, assignIndex)),
			kv("target", assignTarget),
		)
	}
	var mappingObj *vm.Program
	if a.Mapping != "" {
		mappingObj, err = dsh_utils.CompileExpr(a.Mapping)
		if err != nil {
			return nil, errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].assigns[%d].mapping", itemIndex, assignIndex)),
				kv("value", a.Mapping),
			)
		}
	}

	targetsDict[assignTarget] = true
	return newProjectOptionAssignEntity(a.Project, a.Option, a.Mapping, mappingObj), nil
}

// endregion

// region script

type ProjectManifestScript struct {
	Sources        []*ProjectManifestSource
	Imports        []*ProjectManifestImport
	sourceEntities projectSourceEntitySet
	importEntities projectImportEntitySet
}

func NewProjectManifestScript(sources []*ProjectManifestSource, imports []*ProjectManifestImport) *ProjectManifestScript {
	return &ProjectManifestScript{
		Sources: sources,
		Imports: imports,
	}
}

func (s *ProjectManifestScript) init(manifest *ProjectManifest) error {
	sourceEntities := projectSourceEntitySet{}
	for i := 0; i < len(s.Sources); i++ {
		src := s.Sources[i]
		if sourceEntity, err := src.init(manifest, "script", i); err != nil {
			return err
		} else {
			sourceEntities = append(sourceEntities, sourceEntity)
		}
	}

	importEntities := projectImportEntitySet{}
	for i := 0; i < len(s.Imports); i++ {
		imp := s.Imports[i]
		if importEntity, err := imp.init(manifest, "script", i); err != nil {
			return err
		} else {
			importEntities = append(importEntities, importEntity)
		}
	}

	s.sourceEntities = sourceEntities
	s.importEntities = importEntities
	return nil
}

// endregion

// region config

type ProjectManifestConfig struct {
	Sources        []*ProjectManifestSource
	Imports        []*ProjectManifestImport
	sourceEntities projectSourceEntitySet
	importEntities projectImportEntitySet
}

func NewProjectManifestConfig(sources []*ProjectManifestSource, imports []*ProjectManifestImport) *ProjectManifestConfig {
	return &ProjectManifestConfig{
		Sources: sources,
		Imports: imports,
	}
}

func (c *ProjectManifestConfig) init(manifest *ProjectManifest) error {
	sourceEntities := projectSourceEntitySet{}
	for i := 0; i < len(c.Sources); i++ {
		src := c.Sources[i]
		if sourceEntity, err := src.init(manifest, "config", i); err != nil {
			return err
		} else {
			sourceEntities = append(sourceEntities, sourceEntity)
		}
	}

	importEntities := projectImportEntitySet{}
	for i := 0; i < len(c.Imports); i++ {
		imp := c.Imports[i]
		if importEntity, err := imp.init(manifest, "config", i); err != nil {
			return err
		} else {
			importEntities = append(importEntities, importEntity)
		}
	}

	c.sourceEntities = sourceEntities
	c.importEntities = importEntities
	return nil
}

// endregion

// region source

type ProjectManifestSource struct {
	Dir   string
	Files []string
	Match string
}

func NewProjectManifestSource(dir string, files []string, match string) *ProjectManifestSource {
	return &ProjectManifestSource{
		Dir:   dir,
		Files: files,
		Match: match,
	}
}

func (s *ProjectManifestSource) init(manifest *ProjectManifest, scope string, index int) (entity *projectSourceEntity, err error) {
	if s.Dir == "" {
		return nil, errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("%s.sources[%d].dir", scope, index)),
		)
	}
	var matchObj *vm.Program
	if s.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(s.Match)
		if err != nil {
			return nil, errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.sources[%d].match", scope, index)),
				kv("value", s.Match),
			)
		}
	}

	return newProjectSourceEntity(s.Dir, s.Files, s.Match, matchObj), nil
}

// endregion

// region import

type ProjectManifestImport struct {
	Link  string
	Match string
}

func NewProjectManifestImport(link string, match string) *ProjectManifestImport {
	return &ProjectManifestImport{
		Link:  link,
		Match: match,
	}
}

func (i *ProjectManifestImport) init(manifest *ProjectManifest, scope string, index int) (entity *projectImportEntity, err error) {
	if i.Link == "" {
		return nil, errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("%s.imports[%d].link", scope, index)),
		)
	}
	linkObj, err := parseProjectLink(i.Link)
	if err != nil {
		return nil, errW(err, "project manifest invalid",
			reason("value invalid"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("%s.imports[%d].link", scope, index)),
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
				kv("field", fmt.Sprintf("%s.imports[%d].match", scope, index)),
				kv("value", i.Match),
			)
		}
	}

	return newProjectImportEntity(i.Link, i.Match, linkObj, matchObj), nil
}

// endregion
