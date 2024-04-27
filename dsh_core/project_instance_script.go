package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"text/template"
	"time"
)

type ProjectInstanceScript struct {
	SourceContainer *ProjectInstanceScriptSourceContainer
	ImportContainer *ProjectInstanceImportShallowContainer
}

type ProjectInstanceScriptSource struct {
	SourcePath string
	SourceName string
}

type ProjectInstanceScriptSourceContainer struct {
	Context            *Context
	SourceNameMap      map[string]*ProjectInstanceScriptSource
	PlainSources       []*ProjectInstanceScriptSource
	TemplateSources    []*ProjectInstanceScriptSource
	TemplateLibSources []*ProjectInstanceScriptSource
}

func NewProjectInstanceScript(context *Context) *ProjectInstanceScript {
	return &ProjectInstanceScript{
		SourceContainer: &ProjectInstanceScriptSourceContainer{
			Context:       context,
			SourceNameMap: make(map[string]*ProjectInstanceScriptSource),
		},
		ImportContainer: NewShallowImportContainer(context, ProjectInstanceImportScopeScript),
	}
}

func (container *ProjectInstanceScriptSourceContainer) ScanSources(sourceDir string, includeFiles []string) error {
	filePaths, fileTypes, err := dsh_utils.ScanFiles(sourceDir, includeFiles, []dsh_utils.FileType{
		dsh_utils.FileTypePlain,
		dsh_utils.FileTypeTemplate,
		dsh_utils.FileTypeTemplateLib,
	})
	if err != nil {
		return err
	}
	for j := 0; j < len(filePaths); j++ {
		filePath := filePaths[j]
		fileType := fileTypes[j]
		source := &ProjectInstanceScriptSource{
			SourcePath: filepath.Join(sourceDir, filePaths[j]),
			SourceName: filePath,
		}
		if fileType == dsh_utils.FileTypeTemplate {
			source.SourceName = source.SourceName[:len(source.SourceName)-len(".dtpl")]
		}
		if existSource, exist := container.SourceNameMap[source.SourceName]; exist {
			if existSource.SourcePath == source.SourcePath {
				continue
			}
			return dsh_utils.NewError("script source name is duplicated", map[string]any{
				"sourceName":  source.SourceName,
				"sourcePath1": source.SourcePath,
				"sourcePath2": existSource.SourcePath,
			})
		}
		container.SourceNameMap[source.SourceName] = source
		switch fileType {
		case dsh_utils.FileTypePlain:
			container.PlainSources = append(container.PlainSources, source)
		case dsh_utils.FileTypeTemplate:
			container.TemplateSources = append(container.TemplateSources, source)
		case dsh_utils.FileTypeTemplateLib:
			container.TemplateLibSources = append(container.TemplateLibSources, source)
		}
	}
	return nil
}

func (container *ProjectInstanceScriptSourceContainer) BuildSources(config map[string]any, funcs template.FuncMap, outputPath string) (err error) {
	for i := 0; i < len(container.PlainSources); i++ {
		startTime := time.Now()
		source := container.PlainSources[i]
		outputTargetPath := filepath.Join(outputPath, source.SourceName)
		container.Context.Logger.Info("build script start: source=%s, target=%s", source.SourcePath, outputTargetPath)
		err = dsh_utils.LinkOrCopyFile(source.SourcePath, outputTargetPath)
		if err != nil {
			return err
		}
		container.Context.Logger.Info("build script finish: elapsed=%s", time.Since(startTime))
	}

	var templateLibSourcePaths []string
	for i := 0; i < len(container.TemplateLibSources); i++ {
		templateLibSourcePaths = append(templateLibSourcePaths, container.TemplateLibSources[i].SourcePath)
	}
	for i := 0; i < len(container.TemplateSources); i++ {
		startTime := time.Now()
		source := container.TemplateSources[i]
		outputTargetPath := filepath.Join(outputPath, source.SourceName)
		container.Context.Logger.Info("build script start: source=%s, target=%s", source.SourcePath, outputTargetPath)
		if err = BuildTemplate(config, funcs, source.SourcePath, templateLibSourcePaths, outputTargetPath); err != nil {
			return err
		}
		container.Context.Logger.Info("build script finish: elapsed=%s", time.Since(startTime))
	}

	return nil
}
