package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"strings"
	"time"
)

// region projectScriptInstance

type projectScriptInstance struct {
	SourceContainer *projectScriptSourceInstanceContainer
	ImportContainer *projectImportInstanceContainer
}

func newProjectScriptInstance(context *appContext, setting *projectSetting, option *projectOptionInstance) (script *projectScriptInstance, err error) {
	sc, err := newProjectScriptSourceInstanceContainer(context, setting, option)
	if err != nil {
		return nil, err
	}
	ic, err := newProjectImportInstanceContainer(context, setting, option, projectImportScopeScript)
	if err != nil {
		return nil, err
	}
	script = &projectScriptInstance{
		SourceContainer: sc,
		ImportContainer: ic,
	}
	return script, nil
}

// endregion

// region projectScriptSourceInstance

type projectScriptSourceInstance struct {
	SourcePath string
	SourceName string
}

// endregion

// region container

type projectScriptSourceInstanceContainer struct {
	context            *appContext
	ProjectName        string
	PlainSources       []*projectScriptSourceInstance
	TemplateSources    []*projectScriptSourceInstance
	TemplateLibSources []*projectScriptSourceInstance
	sourcesByName      map[string]*projectScriptSourceInstance
}

func newProjectScriptSourceInstanceContainer(context *appContext, settings *projectSetting, option *projectOptionInstance) (*projectScriptSourceInstanceContainer, error) {
	container := &projectScriptSourceInstanceContainer{
		context:       context,
		ProjectName:   settings.Name,
		sourcesByName: map[string]*projectScriptSourceInstance{},
	}
	for i := 0; i < len(settings.ScriptSourceSettings); i++ {
		source := settings.ScriptSourceSettings[i]
		matched, err := option.evaluator.EvalBoolExpr(source.match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if err = container.scanSources(filepath.Join(settings.Path, source.Dir), source.Files); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *projectScriptSourceInstanceContainer) scanSources(sourceDir string, includeFiles []string) error {
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
		source := &projectScriptSourceInstance{
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

func (c *projectScriptSourceInstanceContainer) makeSources(evaluator *Evaluator, outputPath string, useHardLink bool) (targetNames []string, err error) {
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
