package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
)

// region manifest

type projectManifest struct {
	Name         string
	Runtime      *projectManifestRuntime
	Option       *projectManifestOption
	Script       *projectManifestScript
	Config       *projectManifestConfig
	manifestPath string
	entity       *projectEntity
}

func loadProjectManifest(projectPath string) (manifest *projectManifest, err error) {
	manifest = &projectManifest{
		Runtime: &projectManifestRuntime{},
		Option:  &projectManifestOption{},
		Script:  &projectManifestScript{},
		Config:  &projectManifestConfig{},
	}
	metadata, err := dsh_utils.DeserializeByDir(projectPath, []string{"project"}, manifest, true)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("load manifest from dir error"),
			kv("projectPath", projectPath),
		)
	}
	manifest.manifestPath = metadata.Path
	if err = manifest.init(projectPath); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *projectManifest) DescExtraKeyValues() KVS {
	return KVS{
		kv("manifestPath", m.manifestPath),
	}
}

func (m *projectManifest) init(projectPath string) (err error) {
	if m.entity != nil {
		return nil
	}

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
	if err = m.Runtime.init(m); err != nil {
		return err
	}

	var optionDeclares projectOptionDeclareEntitySet
	var optionVerifies projectOptionVerifyEntitySet
	if optionDeclares, optionVerifies, err = m.Option.init(m); err != nil {
		return err
	}

	var scriptSources projectSourceEntitySet
	var scriptImports projectImportEntitySet
	if scriptSources, scriptImports, err = m.Script.init(m); err != nil {
		return err
	}

	var configSources projectSourceEntitySet
	var configImports projectImportEntitySet
	if configSources, configImports, err = m.Config.init(m); err != nil {
		return err
	}

	m.entity = newProjectEntity(m.Name, projectPath, optionDeclares, optionVerifies, scriptSources, scriptImports, configSources, configImports)
	return nil
}

// endregion

// region runtime

type projectManifestRuntime struct {
	MinVersion dsh_utils.Version `yaml:"minVersion" toml:"minVersion" json:"minVersion"`
	MaxVersion dsh_utils.Version `yaml:"maxVersion" toml:"maxVersion" json:"maxVersion"`
}

func (r *projectManifestRuntime) init(manifest *projectManifest) (err error) {
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

type projectManifestOption struct {
	Items    []*projectManifestOptionItem
	Verifies []string
}

type projectManifestOptionItem struct {
	Name     string
	Type     projectOptionValueType
	Choices  []string
	Default  *string
	Optional bool
	Assigns  []*projectManifestOptionItemAssign
}

type projectManifestOptionItemAssign struct {
	Project string
	Option  string
	Mapping string
}

func (o *projectManifestOption) init(manifest *projectManifest) (projectOptionDeclareEntitySet, projectOptionVerifyEntitySet, error) {
	declares := projectOptionDeclareEntitySet{}
	optionNamesDict := map[string]bool{}
	assignTargetsDict := map[string]bool{}
	for i := 0; i < len(o.Items); i++ {
		if declareEntity, err := o.Items[i].init(manifest, optionNamesDict, assignTargetsDict, i); err != nil {
			return nil, nil, err
		} else {
			declares = append(declares, declareEntity)
		}
	}

	verifies := projectOptionVerifyEntitySet{}
	for i := 0; i < len(o.Verifies); i++ {
		expr := o.Verifies[i]
		if expr == "" {
			return nil, nil, errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
			)
		}
		exprObj, err := dsh_utils.CompileExpr(expr)
		if err != nil {
			return nil, nil, errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
				kv("value", expr),
			)
		}
		verifies = append(verifies, newProjectOptionVerifyEntity(expr, exprObj))
	}

	return declares, verifies, nil
}

func (i *projectManifestOptionItem) init(manifest *projectManifest, itemNamesDict, assignTargetsDict map[string]bool, itemIndex int) (entity *projectOptionDeclareEntity, err error) {
	if i.Name == "" {
		return nil, errN("project manifest invalid",
			reason("name empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("option.items[%d].name", itemIndex)),
		)
	}
	if checked := projectOptionNameCheckRegex.MatchString(i.Name); !checked {
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

func (a *projectManifestOptionItemAssign) init(manifest *projectManifest, targetsDict map[string]bool, itemIndex int, assignIndex int) (entity *projectOptionAssignEntity, err error) {
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
	var mappingObj *EvalExpr
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

type projectManifestScript struct {
	Sources []*projectManifestSource
	Imports []*projectManifestImport
}

func (s *projectManifestScript) init(manifest *projectManifest) (projectSourceEntitySet, projectImportEntitySet, error) {
	sourceEntities := projectSourceEntitySet{}
	for i := 0; i < len(s.Sources); i++ {
		src := s.Sources[i]
		if sourceEntity, err := src.init(manifest, "script", i); err != nil {
			return nil, nil, err
		} else {
			sourceEntities = append(sourceEntities, sourceEntity)
		}
	}

	importEntities := projectImportEntitySet{}
	for i := 0; i < len(s.Imports); i++ {
		imp := s.Imports[i]
		if importEntity, err := imp.init(manifest, "script", i); err != nil {
			return nil, nil, err
		} else {
			importEntities = append(importEntities, importEntity)
		}
	}

	return sourceEntities, importEntities, nil
}

// endregion

// region config

type projectManifestConfig struct {
	Sources []*projectManifestSource
	Imports []*projectManifestImport
}

func (c *projectManifestConfig) init(manifest *projectManifest) (projectSourceEntitySet, projectImportEntitySet, error) {
	sources := projectSourceEntitySet{}
	for i := 0; i < len(c.Sources); i++ {
		src := c.Sources[i]
		if sourceEntity, err := src.init(manifest, "config", i); err != nil {
			return nil, nil, err
		} else {
			sources = append(sources, sourceEntity)
		}
	}

	imports := projectImportEntitySet{}
	for i := 0; i < len(c.Imports); i++ {
		imp := c.Imports[i]
		if importEntity, err := imp.init(manifest, "config", i); err != nil {
			return nil, nil, err
		} else {
			imports = append(imports, importEntity)
		}
	}

	return sources, imports, nil
}

// endregion

// region source

type projectManifestSource struct {
	Dir   string
	Files []string
	Match string
}

func (s *projectManifestSource) init(manifest *projectManifest, scope string, index int) (entity *projectSourceEntity, err error) {
	if s.Dir == "" {
		return nil, errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("%s.sources[%d].dir", scope, index)),
		)
	}
	var matchObj *EvalExpr
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

type projectManifestImport struct {
	Link  string
	Match string
}

func (i *projectManifestImport) init(manifest *projectManifest, scope string, index int) (entity *projectImportEntity, err error) {
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

	var matchObj *EvalExpr
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
