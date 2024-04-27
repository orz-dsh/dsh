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
	plainSourcePaths, templateSourcePaths, templateLibSourcePaths, err := dsh_utils.ScanScriptSources(sourceDir, includeFiles)
	if err != nil {
		return err
	}
	for i := 0; i < len(plainSourcePaths); i++ {
		source := &ProjectInstanceScriptSource{
			SourcePath: filepath.Join(sourceDir, plainSourcePaths[i]),
			SourceName: plainSourcePaths[i],
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
		container.PlainSources = append(container.PlainSources, source)
	}
	for i := 0; i < len(templateSourcePaths); i++ {
		source := &ProjectInstanceScriptSource{
			SourcePath: filepath.Join(sourceDir, templateSourcePaths[i]),
			SourceName: templateSourcePaths[i][:len(templateSourcePaths[i])-len(".dtpl")],
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
		container.TemplateSources = append(container.TemplateSources, source)
	}
	for i := 0; i < len(templateLibSourcePaths); i++ {
		source := &ProjectInstanceScriptSource{
			SourcePath: filepath.Join(sourceDir, templateLibSourcePaths[i]),
			SourceName: templateLibSourcePaths[i],
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
		container.TemplateLibSources = append(container.TemplateLibSources, source)
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
