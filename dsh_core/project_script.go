package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type projectScript struct {
	sourceContainer *projectScriptSourceContainer
	importContainer *projectImportContainer
}

type projectScriptSource struct {
	sourcePath string
	sourceName string
}

type projectScriptSourceContainer struct {
	context            *appContext
	manifest           *projectManifest
	plainSources       []*projectScriptSource
	templateSources    []*projectScriptSource
	templateLibSources []*projectScriptSource
	sourcesByName      map[string]*projectScriptSource
}

func loadProjectScript(context *appContext, manifest *projectManifest) (ps *projectScript, err error) {
	sc, err := loadProjectScriptSourceContainer(context, manifest)
	if err != nil {
		return nil, err
	}
	ic, err := loadProjectImportContainer(context, manifest, projectImportScopeScript)
	if err != nil {
		return nil, err
	}
	ps = &projectScript{
		sourceContainer: sc,
		importContainer: ic,
	}
	return ps, nil
}

func loadProjectScriptSourceContainer(context *appContext, manifest *projectManifest) (container *projectScriptSourceContainer, err error) {
	container = &projectScriptSourceContainer{
		context:       context,
		manifest:      manifest,
		sourcesByName: make(map[string]*projectScriptSource),
	}
	for i := 0; i < len(manifest.Script.Sources); i++ {
		src := manifest.Script.Sources[i]
		if src.Dir != "" {
			if src.Match != "" {
				matched, err := context.option.evalProjectMatchExpr(manifest, src.match)
				if err != nil {
					return nil, err
				}
				if !matched {
					continue
				}
			}
			if err = container.scanSources(filepath.Join(manifest.projectPath, src.Dir), src.Files); err != nil {
				return nil, err
			}
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
			sourcePath: filepath.Join(sourceDir, filePaths[j]),
			sourceName: filePath,
		}
		if fileType == dsh_utils.FileTypeTemplate {
			source.sourceName = source.sourceName[:len(source.sourceName)-len(".dtpl")]
		}
		if existSource, exist := c.sourcesByName[source.sourceName]; exist {
			if existSource.sourcePath == source.sourcePath {
				continue
			}
			return errN("scan script sources error",
				reason("source name duplicated"),
				kv("sourceName", source.sourceName),
				kv("sourcePath1", source.sourcePath),
				kv("sourcePath2", existSource.sourcePath),
			)
		}
		switch fileType {
		case dsh_utils.FileTypePlain:
			c.plainSources = append(c.plainSources, source)
		case dsh_utils.FileTypeTemplate:
			c.templateSources = append(c.templateSources, source)
		case dsh_utils.FileTypeTemplateLib:
			c.templateLibSources = append(c.templateLibSources, source)
		default:
			// impossible
			panic(desc("script source type unsupported",
				kv("filePath", filePath),
				kv("fileType", fileType),
			))
		}
		c.sourcesByName[source.sourceName] = source
	}
	return nil
}

func (c *projectScriptSourceContainer) makeSources(data map[string]any, funcs template.FuncMap, outputPath string) (targetNames []string, err error) {
	for i := 0; i < len(c.plainSources); i++ {
		startTime := time.Now()
		source := c.plainSources[i]
		target := filepath.Join(c.manifest.Name, source.sourceName)
		targetPath := filepath.Join(outputPath, target)
		c.context.logger.InfoDesc("make script sources start",
			kv("sourceType", dsh_utils.FileTypePlain),
			kv("sourcePath", source.sourcePath),
			kv("targetPath", targetPath),
		)
		err = dsh_utils.LinkOrCopyFile(source.sourcePath, targetPath)
		if err != nil {
			return nil, errW(err, "make script sources error",
				reason("link or copy file error"),
				kv("sourceType", dsh_utils.FileTypePlain),
				kv("sourcePath", source.sourcePath),
				kv("targetPath", targetPath),
			)
		}
		targetNames = append(targetNames, strings.ReplaceAll(target, "\\", "/"))
		c.context.logger.InfoDesc("make script sources finish",
			kv("elapsed", time.Since(startTime)),
		)
	}
	var templateLibSourcePaths []string
	for i := 0; i < len(c.templateLibSources); i++ {
		templateLibSourcePaths = append(templateLibSourcePaths, c.templateLibSources[i].sourcePath)
	}
	for i := 0; i < len(c.templateSources); i++ {
		startTime := time.Now()
		source := c.templateSources[i]
		target := filepath.Join(c.manifest.Name, source.sourceName)
		targetPath := filepath.Join(outputPath, target)
		c.context.logger.InfoDesc("make script sources start",
			kv("sourceType", dsh_utils.FileTypeTemplate),
			kv("sourcePath", source.sourcePath),
			kv("targetPath", targetPath),
		)
		if err = executeFileTemplate(source.sourcePath, templateLibSourcePaths, targetPath, data, funcs); err != nil {
			return nil, errW(err, "make script sources error",
				reason("make template error"),
				kv("sourceType", dsh_utils.FileTypeTemplate),
				kv("sourcePath", source.sourcePath),
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
