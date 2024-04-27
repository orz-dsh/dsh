package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"text/template"
	"time"
)

type projectInstanceScript struct {
	sourceContainer *projectInstanceScriptSourceContainer
	importContainer *projectInstanceImportShallowContainer
}

type projectInstanceScriptSource struct {
	sourcePath string
	sourceName string
}

type projectInstanceScriptSourceContainer struct {
	context            *Context
	sourceNameMap      map[string]*projectInstanceScriptSource
	plainSources       []*projectInstanceScriptSource
	templateSources    []*projectInstanceScriptSource
	templateLibSources []*projectInstanceScriptSource
}

func newProjectInstanceScript(context *Context) *projectInstanceScript {
	return &projectInstanceScript{
		sourceContainer: &projectInstanceScriptSourceContainer{
			context:       context,
			sourceNameMap: make(map[string]*projectInstanceScriptSource),
		},
		importContainer: newProjectInstanceImportShallowContainer(context, projectInstanceImportScopeScript),
	}
}

func (container *projectInstanceScriptSourceContainer) scanSources(sourceDir string, includeFiles []string) error {
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
		source := &projectInstanceScriptSource{
			sourcePath: filepath.Join(sourceDir, filePaths[j]),
			sourceName: filePath,
		}
		if fileType == dsh_utils.FileTypeTemplate {
			source.sourceName = source.sourceName[:len(source.sourceName)-len(".dtpl")]
		}
		if existSource, exist := container.sourceNameMap[source.sourceName]; exist {
			if existSource.sourcePath == source.sourcePath {
				continue
			}
			return dsh_utils.NewError("script source name is duplicated", map[string]any{
				"sourceName":  source.sourceName,
				"sourcePath1": source.sourcePath,
				"sourcePath2": existSource.sourcePath,
			})
		}
		container.sourceNameMap[source.sourceName] = source
		switch fileType {
		case dsh_utils.FileTypePlain:
			container.plainSources = append(container.plainSources, source)
		case dsh_utils.FileTypeTemplate:
			container.templateSources = append(container.templateSources, source)
		case dsh_utils.FileTypeTemplateLib:
			container.templateLibSources = append(container.templateLibSources, source)
		}
	}
	return nil
}

func (container *projectInstanceScriptSourceContainer) make(config map[string]any, funcs template.FuncMap, outputPath string) (err error) {
	for i := 0; i < len(container.plainSources); i++ {
		startTime := time.Now()
		source := container.plainSources[i]
		outputTargetPath := filepath.Join(outputPath, source.sourceName)
		container.context.Logger.Info("make file start: source=%s, target=%s", source.sourcePath, outputTargetPath)
		err = dsh_utils.LinkOrCopyFile(source.sourcePath, outputTargetPath)
		if err != nil {
			return err
		}
		container.context.Logger.Info("make file finish: elapsed=%s", time.Since(startTime))
	}

	var templateLibSourcePaths []string
	for i := 0; i < len(container.templateLibSources); i++ {
		templateLibSourcePaths = append(templateLibSourcePaths, container.templateLibSources[i].sourcePath)
	}
	for i := 0; i < len(container.templateSources); i++ {
		startTime := time.Now()
		source := container.templateSources[i]
		outputTargetPath := filepath.Join(outputPath, source.sourceName)
		container.context.Logger.Info("make file start: source=%s, target=%s", source.sourcePath, outputTargetPath)
		if err = makeTemplate(config, funcs, source.sourcePath, templateLibSourcePaths, outputTargetPath); err != nil {
			return err
		}
		container.context.Logger.Info("make file finish: elapsed=%s", time.Since(startTime))
	}

	return nil
}
