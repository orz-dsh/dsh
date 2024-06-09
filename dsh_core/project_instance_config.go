package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
)

// region projectConfigInstance

type projectConfigInstance struct {
	SourceContainer *projectConfigSourceInstanceContainer
	ImportContainer *projectImportInstanceContainer
}

func newProjectConfigInstance(context *appContext, setting *projectSetting, option *projectOptionInstance) (instance *projectConfigInstance, err error) {
	sc, err := newProjectConfigSourceInstanceContainer(context, setting, option)
	if err != nil {
		return nil, err
	}
	ic, err := newProjectImportInstanceContainer(context, setting, option, projectImportScopeConfig)
	if err != nil {
		return nil, err
	}
	instance = &projectConfigInstance{
		SourceContainer: sc,
		ImportContainer: ic,
	}
	return instance, nil
}

func (i *projectConfigInstance) inspect() *ProjectConfigInstanceInspection {
	sources := i.SourceContainer.inspect()
	imports := i.ImportContainer.inspect()
	return newProjectConfigInstanceInspection(sources, imports)
}

// endregion

// region projectConfigSourceInstance

type projectConfigSourceInstance struct {
	SourcePath   string
	SourceFormat projectConfigSourceFormat
	content      *projectConfigContentInstance
}

type projectConfigSourceFormat = dsh_utils.SerializationFormat

func (i *projectConfigSourceInstance) loadContent() error {
	if i.content == nil {
		if content, err := newProjectConfigContentInstance(i.SourcePath, i.SourceFormat); err != nil {
			return err
		} else {
			i.content = content
		}
	}
	return nil
}

func (i *projectConfigSourceInstance) inspect() *ProjectConfigSourceInstanceInspection {
	return newProjectConfigSourceInstanceInspection(i.SourcePath)
}

// endregion

// region projectConfigContentInstance

type projectConfigContentInstance struct {
	Order        int64
	Merges       map[string]projectConfigContentMergeMode
	Configs      map[string]any
	sourcePath   string
	sourceFormat projectConfigSourceFormat
}

type projectConfigContentMergeMode = dsh_utils.MapMergeMode

func newProjectConfigContentInstance(sourcePath string, sourceFormat projectConfigSourceFormat) (*projectConfigContentInstance, error) {
	content := &projectConfigContentInstance{
		Merges:       map[string]projectConfigContentMergeMode{},
		sourcePath:   sourcePath,
		sourceFormat: sourceFormat,
	}
	if _, err := dsh_utils.DeserializeFromFile(sourcePath, sourceFormat, content); err != nil {
		return nil, errW(err, "load config sources error",
			reason("deserialize error"),
			kv("sourcePath", sourceFormat),
			kv("sourceFormat", sourcePath),
		)
	}
	for k, v := range content.Merges {
		switch v {
		case dsh_utils.MapMergeModeReplace:
		case dsh_utils.MapMergeModeInsert:
		default:
			return nil, errN("load config sources error",
				reason("merge mode invalid"),
				kv("file", sourcePath),
				kv("field", fmt.Sprintf("merges[%s]", k)),
				kv("value", v),
			)
		}
	}
	return content, nil
}

func (i *projectConfigContentInstance) merge(configs map[string]any, configsTraces map[string]any) error {
	if _, _, err := dsh_utils.MapMerge(configs, i.Configs, i.Merges, i.sourcePath, configsTraces); err != nil {
		return errW(err, "merge config content error",
			kv("sourcePath", i.sourcePath),
		)
	}
	return nil
}

// endregion

// region projectConfigSourceInstanceContainer

type projectConfigSourceInstanceContainer struct {
	context         *appContext
	Sources         []*projectConfigSourceInstance
	sourcePathsDict map[string]bool
}

func newProjectConfigSourceInstanceContainer(context *appContext, setting *projectSetting, option *projectOptionInstance) (*projectConfigSourceInstanceContainer, error) {
	container := &projectConfigSourceInstanceContainer{
		context:         context,
		sourcePathsDict: map[string]bool{},
	}
	for i := 0; i < len(setting.ConfigSourceSettings); i++ {
		source := setting.ConfigSourceSettings[i]
		matched, err := option.evaluator.EvalBoolExpr(source.match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if err = container.scanSources(filepath.Join(setting.Path, source.Dir), source.Files); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *projectConfigSourceInstanceContainer) scanSources(sourceDir string, includeFiles []string) error {
	files, err := dsh_utils.ScanFiles(sourceDir, includeFiles, []dsh_utils.FileType{
		dsh_utils.FileTypeYaml,
		dsh_utils.FileTypeToml,
		dsh_utils.FileTypeJson,
	})
	if err != nil {
		return err
	}
	for i := 0; i < len(files); i++ {
		file := files[i]
		source := &projectConfigSourceInstance{
			SourcePath:   file.Path,
			SourceFormat: dsh_utils.GetSerializationFormat(file.Type),
		}
		if c.sourcePathsDict[source.SourcePath] {
			continue
		}
		c.Sources = append(c.Sources, source)
		c.sourcePathsDict[source.SourcePath] = true
	}
	return nil
}

func (c *projectConfigSourceInstanceContainer) loadContents() (contents []*projectConfigContentInstance, err error) {
	for i := 0; i < len(c.Sources); i++ {
		source := c.Sources[i]
		err = source.loadContent()
		if err != nil {
			return nil, err
		}
		contents = append(contents, source.content)
	}
	return contents, nil
}

func (c *projectConfigSourceInstanceContainer) inspect() (sources []*ProjectConfigSourceInstanceInspection) {
	for i := 0; i < len(c.Sources); i++ {
		sources = append(sources, c.Sources[i].inspect())
	}
	return sources
}

// endregion

// region ProjectConfigInstanceInspection

type ProjectConfigInstanceInspection struct {
	Sources []*ProjectConfigSourceInstanceInspection `yaml:"sources,omitempty" toml:"sources,omitempty" json:"sources,omitempty"`
	Imports []*ProjectImportInstanceInspection       `yaml:"imports,omitempty" toml:"imports,omitempty" json:"imports,omitempty"`
}

func newProjectConfigInstanceInspection(sources []*ProjectConfigSourceInstanceInspection, imports []*ProjectImportInstanceInspection) *ProjectConfigInstanceInspection {
	return &ProjectConfigInstanceInspection{
		Sources: sources,
		Imports: imports,
	}
}

// endregion

// region ProjectConfigSourceInstanceInspection

type ProjectConfigSourceInstanceInspection struct {
	SourcePath string `yaml:"sourcePath" toml:"sourcePath" json:"sourcePath"`
}

func newProjectConfigSourceInstanceInspection(sourcePath string) *ProjectConfigSourceInstanceInspection {
	return &ProjectConfigSourceInstanceInspection{
		SourcePath: sourcePath,
	}
}

// endregion
