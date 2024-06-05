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

func newProjectConfigInstance(context *appContext, setting *projectSetting, option *projectOption) (instance *projectConfigInstance, err error) {
	sc, err := newProjectConfigSourceInstanceContainer(context, setting, option)
	if err != nil {
		return nil, err
	}
	ic, err := makeProjectImportContainer(context, setting, option, projectImportScopeConfig)
	if err != nil {
		return nil, err
	}
	instance = &projectConfigInstance{
		SourceContainer: sc,
		ImportContainer: ic,
	}
	return instance, nil
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
	if i.content != nil {
		return nil
	}
	content, err := newProjectConfigContentInstance(i.SourcePath, i.SourceFormat)
	if err != nil {
		return err
	}
	i.content = content
	return nil
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

func (i *projectConfigContentInstance) merge(target map[string]any) error {
	if _, err := dsh_utils.MergeMap(target, i.Configs, i.Merges); err != nil {
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

func newProjectConfigSourceInstanceContainer(context *appContext, setting *projectSetting, option *projectOption) (*projectConfigSourceInstanceContainer, error) {
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

// endregion
