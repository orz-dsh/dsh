package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
)

// region manifest

type projectManifest struct {
	Name         string
	Runtime      *projectManifestRuntime
	Option       *projectPrefOption
	Script       *projectManifestScript
	Config       *projectManifestConfig
	manifestPath string
	entity       *projectSchema
}

func loadProjectManifest(projectPath string) (manifest *projectManifest, err error) {
	manifest = &projectManifest{
		Runtime: &projectManifestRuntime{},
		Option:  &projectPrefOption{},
		Script:  &projectManifestScript{},
		Config:  &projectManifestConfig{},
	}
	metadata, err := dsh_utils.DeserializeFromDir(projectPath, []string{"project"}, manifest, true)
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
		kv("path", m.manifestPath),
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

	var optionDeclares projectSchemaOptionSet
	var optionVerifies projectSchemaOptionVerifySet
	if optionDeclares, optionVerifies, err = m.Option.init(m); err != nil {
		return err
	}

	var scriptSources projectSchemaSourceSet
	var scriptImports projectSchemaImportSet
	if scriptSources, scriptImports, err = m.Script.init(m); err != nil {
		return err
	}

	var configSources projectSchemaSourceSet
	var configImports projectSchemaImportSet
	if configSources, configImports, err = m.Config.init(m); err != nil {
		return err
	}

	m.entity = newProjectSchema(m.Name, projectPath, optionDeclares, optionVerifies, scriptSources, scriptImports, configSources, configImports)
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

// region script

type projectManifestScript struct {
	Sources []*projectManifestSource
	Imports []*projectManifestImport
}

func (s *projectManifestScript) init(manifest *projectManifest) (projectSchemaSourceSet, projectSchemaImportSet, error) {
	sourceEntities := projectSchemaSourceSet{}
	for i := 0; i < len(s.Sources); i++ {
		src := s.Sources[i]
		if sourceEntity, err := src.init(manifest, "script", i); err != nil {
			return nil, nil, err
		} else {
			sourceEntities = append(sourceEntities, sourceEntity)
		}
	}

	importEntities := projectSchemaImportSet{}
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

func (c *projectManifestConfig) init(manifest *projectManifest) (projectSchemaSourceSet, projectSchemaImportSet, error) {
	sources := projectSchemaSourceSet{}
	for i := 0; i < len(c.Sources); i++ {
		src := c.Sources[i]
		if sourceEntity, err := src.init(manifest, "config", i); err != nil {
			return nil, nil, err
		} else {
			sources = append(sources, sourceEntity)
		}
	}

	imports := projectSchemaImportSet{}
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

func (s *projectManifestSource) init(manifest *projectManifest, scope string, index int) (entity *projectSchemaSource, err error) {
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

	return newProjectSchemaSource(s.Dir, s.Files, s.Match, matchObj), nil
}

// endregion

// region import

type projectManifestImport struct {
	Link  string
	Match string
}

func (i *projectManifestImport) init(manifest *projectManifest, scope string, index int) (entity *projectSchemaImport, err error) {
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

	return newProjectSchemaImport(i.Link, i.Match, linkObj, matchObj), nil
}

// endregion
