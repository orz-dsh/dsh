package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"strings"
	"time"
)

// region script

type projectScript struct {
	SourceContainer *projectScriptSourceContainer
	ImportContainer *projectImportContainer
}

func makeProjectScript(context *appContext, entity *projectSetting, option *projectOption) (script *projectScript, err error) {
	sc, err := makeProjectScriptSourceContainer(context, entity, option)
	if err != nil {
		return nil, err
	}
	ic, err := makeProjectImportContainer(context, entity, option, projectImportScopeScript)
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
	ProjectName        string
	PlainSources       []*projectScriptSource
	TemplateSources    []*projectScriptSource
	TemplateLibSources []*projectScriptSource
	sourcesByName      map[string]*projectScriptSource
}

func makeProjectScriptSourceContainer(context *appContext, entity *projectSetting, option *projectOption) (container *projectScriptSourceContainer, err error) {
	container = &projectScriptSourceContainer{
		context:       context,
		ProjectName:   entity.Name,
		sourcesByName: map[string]*projectScriptSource{},
	}
	sources := entity.ScriptSourceSettings
	for i := 0; i < len(sources); i++ {
		source := sources[i]
		matched, err := option.evaluator.EvalBoolExpr(source.match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if err = container.scanSources(filepath.Join(entity.Path, source.Dir), source.Files); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *projectScriptSourceContainer) scanSources(sourceDir string, includeFiles []string) error {
	files, err := dsh_utils.ScanFiles(sourceDir, includeFiles, []dsh_utils.FileType{
		dsh_utils.FileTypePlain,
		dsh_utils.FileTypeTemplate,
		dsh_utils.FileTypeTemplateLib,
	})
	if err != nil {
		return err
	}
	for j := 0; j < len(files); j++ {
		file := files[j]
		source := &projectScriptSource{
			SourcePath: file.Path,
			SourceName: file.RelPath,
		}
		if file.Type == dsh_utils.FileTypeTemplate {
			source.SourceName = source.SourceName[:len(source.SourceName)-len(".dtpl")]
		}
		if existSource, exist := c.sourcesByName[source.SourceName]; exist {
			if existSource.SourcePath == source.SourcePath {
				continue
			}
			return errN("scan script sources error",
				reason("source name duplicated"),
				kv("source", source),
				kv("existSource", existSource),
			)
		}
		switch file.Type {
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

func (c *projectScriptSourceContainer) makeSources(evaluator *Evaluator, outputPath string, useHardLink bool) (targetNames []string, err error) {
	for i := 0; i < len(c.PlainSources); i++ {
		startTime := time.Now()
		source := c.PlainSources[i]
		target := filepath.Join(c.ProjectName, source.SourceName)
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
		target := filepath.Join(c.ProjectName, source.SourceName)
		targetPath := filepath.Join(outputPath, target)
		c.context.logger.InfoDesc("make script sources start",
			kv("sourceType", dsh_utils.FileTypeTemplate),
			kv("sourcePath", source.SourcePath),
			kv("targetPath", targetPath),
		)
		if err = evaluator.EvalFileTemplate(source.SourcePath, templateLibSourcePaths, targetPath); err != nil {
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
