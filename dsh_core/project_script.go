package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
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

func (sc *projectScriptSourceContainer) scanSources(sourceDir string, includeFiles []string) error {
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
		if existSource, exist := sc.sourcesByName[source.sourceName]; exist {
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
			sc.plainSources = append(sc.plainSources, source)
		case dsh_utils.FileTypeTemplate:
			sc.templateSources = append(sc.templateSources, source)
		case dsh_utils.FileTypeTemplateLib:
			sc.templateLibSources = append(sc.templateLibSources, source)
		default:
			// impossible
			panic(desc("script source type unsupported",
				kv("filePath", filePath),
				kv("fileType", fileType),
			))
		}
		sc.sourcesByName[source.sourceName] = source
	}
	return nil
}

func (sc *projectScriptSourceContainer) makeSources(env map[string]any, funcs template.FuncMap, outputPath string) (err error) {
	for i := 0; i < len(sc.plainSources); i++ {
		startTime := time.Now()
		source := sc.plainSources[i]
		outputTargetPath := filepath.Join(outputPath, source.sourceName)
		sc.context.logger.InfoDesc("make script sources start",
			kv("sourceType", dsh_utils.FileTypePlain),
			kv("sourcePath", source.sourcePath),
			kv("targetPath", outputTargetPath),
		)
		err = dsh_utils.LinkOrCopyFile(source.sourcePath, outputTargetPath)
		if err != nil {
			return errW(err, "make script sources error",
				reason("link or copy file error"),
				kv("sourceType", dsh_utils.FileTypePlain),
				kv("sourcePath", source.sourcePath),
				kv("targetPath", outputTargetPath),
			)
		}
		sc.context.logger.InfoDesc("make script sources finish",
			kv("elapsed", time.Since(startTime)),
		)
	}
	var templateLibSourcePaths []string
	for i := 0; i < len(sc.templateLibSources); i++ {
		templateLibSourcePaths = append(templateLibSourcePaths, sc.templateLibSources[i].sourcePath)
	}
	for i := 0; i < len(sc.templateSources); i++ {
		startTime := time.Now()
		source := sc.templateSources[i]
		outputTargetPath := filepath.Join(outputPath, source.sourceName)
		sc.context.logger.InfoDesc("make script sources start",
			kv("sourceType", dsh_utils.FileTypeTemplate),
			kv("sourcePath", source.sourcePath),
			kv("targetPath", outputTargetPath),
		)
		if err = makeTemplate(env, funcs, source.sourcePath, templateLibSourcePaths, outputTargetPath); err != nil {
			return errW(err, "make script sources error",
				reason("make template error"),
				kv("sourceType", dsh_utils.FileTypeTemplate),
				kv("sourcePath", source.sourcePath),
				kv("targetPath", outputTargetPath),
			)
		}
		sc.context.logger.InfoDesc("make script sources finish",
			kv("elapsed", time.Since(startTime)),
		)
	}
	return nil
}
