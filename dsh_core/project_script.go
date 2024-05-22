package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// region script

type projectScript struct {
	SourceContainer *projectScriptSourceContainer
	ImportContainer *projectImportContainer
}

func loadProjectScript(context *appContext, manifest *projectManifest) (script *projectScript, err error) {
	sc, err := loadProjectScriptSourceContainer(context, manifest)
	if err != nil {
		return nil, err
	}
	ic, err := makeProjectImportContainer(context, manifest, projectImportScopeScript)
	if err != nil {
		return nil, err
	}
	script = &projectScript{
		SourceContainer: sc,
		ImportContainer: ic,
	}
	return script, nil
}

// endregion

// region source

type projectScriptSource struct {
	SourcePath string
	SourceName string
}

// endregion

// region container

type projectScriptSourceContainer struct {
	context            *appContext
	manifest           *projectManifest
	PlainSources       []*projectScriptSource
	TemplateSources    []*projectScriptSource
	TemplateLibSources []*projectScriptSource
	sourcesByName      map[string]*projectScriptSource
}

func loadProjectScriptSourceContainer(context *appContext, manifest *projectManifest) (container *projectScriptSourceContainer, err error) {
	container = &projectScriptSourceContainer{
		context:       context,
		manifest:      manifest,
		sourcesByName: make(map[string]*projectScriptSource),
	}
	definitions := manifest.Script.sourceDefinitions
	if context.isMainProject(manifest) {
		definitions = append(definitions, context.Profile.getScriptSourceDefinitions()...)
	}
	for i := 0; i < len(definitions); i++ {
		definition := definitions[i]
		matched, err := context.Option.evalMatch(manifest, definition.match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if err = container.scanSources(filepath.Join(manifest.projectPath, definition.Dir), definition.Files); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *projectScriptSourceContainer) scanSources(sourceDir string, includeFiles []string) error {
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
		source := &projectScriptSource{
			SourcePath: filepath.Join(sourceDir, filePaths[j]),
			SourceName: filePath,
		}
		if fileType == dsh_utils.FileTypeTemplate {
			source.SourceName = source.SourceName[:len(source.SourceName)-len(".dtpl")]
		}
		if existSource, exist := c.sourcesByName[source.SourceName]; exist {
			if existSource.SourcePath == source.SourcePath {
				continue
			}
			return errN("scan script sources error",
				reason("source name duplicated"),
				kv("source1", existSource),
				kv("source2", source),
			)
		}
		switch fileType {
		case dsh_utils.FileTypePlain:
			c.PlainSources = append(c.PlainSources, source)
		case dsh_utils.FileTypeTemplate:
			c.TemplateSources = append(c.TemplateSources, source)
		case dsh_utils.FileTypeTemplateLib:
			c.TemplateLibSources = append(c.TemplateLibSources, source)
		default:
			impossible()
		}
		c.sourcesByName[source.SourceName] = source
	}
	return nil
}

func (c *projectScriptSourceContainer) makeSources(data map[string]any, funcs template.FuncMap, outputPath string, useHardLink bool) (targetNames []string, err error) {
	for i := 0; i < len(c.PlainSources); i++ {
		startTime := time.Now()
		source := c.PlainSources[i]
		target := filepath.Join(c.manifest.Name, source.SourceName)
		targetPath := filepath.Join(outputPath, target)
		c.context.logger.InfoDesc("make script sources start",
			kv("sourceType", dsh_utils.FileTypePlain),
			kv("sourcePath", source.SourcePath),
			kv("targetPath", targetPath),
		)
		if useHardLink {
			err = dsh_utils.LinkOrCopyFile(source.SourcePath, targetPath)
			if err != nil {
				return nil, errW(err, "make script sources error",
					reason("link or copy file error"),
					kv("sourceType", dsh_utils.FileTypePlain),
					kv("sourcePath", source.SourcePath),
					kv("targetPath", targetPath),
				)
			}
		} else {
			err = dsh_utils.CopyFile(source.SourcePath, targetPath)
			if err != nil {
				return nil, errW(err, "make script sources error",
					reason("copy file error"),
					kv("sourceType", dsh_utils.FileTypePlain),
					kv("sourcePath", source.SourcePath),
					kv("targetPath", targetPath),
				)
			}
		}
		targetNames = append(targetNames, strings.ReplaceAll(target, "\\", "/"))
		c.context.logger.InfoDesc("make script sources finish",
			kv("elapsed", time.Since(startTime)),
		)
	}
	var templateLibSourcePaths []string
	for i := 0; i < len(c.TemplateLibSources); i++ {
		templateLibSourcePaths = append(templateLibSourcePaths, c.TemplateLibSources[i].SourcePath)
	}
	for i := 0; i < len(c.TemplateSources); i++ {
		startTime := time.Now()
		source := c.TemplateSources[i]
		target := filepath.Join(c.manifest.Name, source.SourceName)
		targetPath := filepath.Join(outputPath, target)
		c.context.logger.InfoDesc("make script sources start",
			kv("sourceType", dsh_utils.FileTypeTemplate),
			kv("sourcePath", source.SourcePath),
			kv("targetPath", targetPath),
		)
		if err = dsh_utils.EvalFileTemplate(source.SourcePath, templateLibSourcePaths, targetPath, data, funcs); err != nil {
			return nil, errW(err, "make script sources error",
				reason("make template error"),
				kv("sourceType", dsh_utils.FileTypeTemplate),
				kv("sourcePath", source.SourcePath),
				kv("targetPath", targetPath),
			)
		}
		targetNames = append(targetNames, strings.ReplaceAll(target, "\\", "/"))
		c.context.logger.InfoDesc("make script sources finish",
			kv("elapsed", time.Since(startTime)),
		)
	}
	return targetNames, nil
}

// endregion
