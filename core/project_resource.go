package core

import (
	"fmt"
	"github.com/orz-dsh/dsh/utils"
	"path/filepath"
	"strings"
	"time"
)

// region base

type projectResourceConfigFormat = utils.SerializationFormat

type projectResourceConfigMergeMode = utils.MapMergeMode

// endregion

// region projectResourceEntity

type projectResourceEntity struct {
	context          *appContext
	ConfigItems      []*projectResourceConfigItemEntity
	TemplateItems    []*projectResourceTemplateItemEntity
	TemplateLibItems []*projectResourceTemplateLibItemEntity
	PlainItems       []*projectResourcePlainItemEntity
}

func newProjectResourceEntity(context *appContext, setting *projectSetting, option *projectOptionEntity) (*projectResourceEntity, error) {
	resource := &projectResourceEntity{context: context}
	configItemsDict := map[string]bool{}
	templateLibItemsDict := map[string]bool{}
	filesByTarget := map[string]string{}
	for i := 0; i < len(setting.Resource.Items); i++ {
		item := setting.Resource.Items[i]
		matched, err := option.evaluator.EvalBoolExpr(item.match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		dir := filepath.Join(setting.Dir, item.Dir)
		if err = resource.scan(setting.Name, dir, item.Includes, item.Excludes, configItemsDict, templateLibItemsDict, filesByTarget); err != nil {
			return nil, err
		}
	}
	return resource, nil
}

func (e *projectResourceEntity) scan(projectName, dir string, includes, excludes []string, configItemsDict, templateLibItemsDict map[string]bool, filesByTarget map[string]string) error {
	files, err := utils.ScanFiles(dir, includes, excludes, []utils.FileType{
		utils.FileTypeConfigYaml,
		utils.FileTypeConfigToml,
		utils.FileTypeConfigJson,
		utils.FileTypeTemplate,
		utils.FileTypeTemplateLib,
		utils.FileTypePlain,
	})
	if err != nil {
		return err
	}
	for i := 0; i < len(files); i++ {
		file := files[i]
		switch file.Type {
		case utils.FileTypeConfigYaml, utils.FileTypeConfigToml, utils.FileTypeConfigJson:
			if !configItemsDict[file.Path] {
				configItem := &projectResourceConfigItemEntity{
					File:   file.Path,
					Format: utils.GetSerializationFormat(file.Type),
				}
				e.ConfigItems = append(e.ConfigItems, configItem)
				configItemsDict[file.Path] = true
			}
			continue
		case utils.FileTypeTemplateLib:
			if !templateLibItemsDict[file.Path] {
				templateLibItem := &projectResourceTemplateLibItemEntity{
					File: file.Path,
				}
				e.TemplateLibItems = append(e.TemplateLibItems, templateLibItem)
				templateLibItemsDict[file.Path] = true
			}
			continue
		}
		target := filepath.Join(projectName, file.RelPath)
		if file.Type == utils.FileTypeTemplate {
			target = target[:len(target)-len(".dtpl")]
		}
		if existFile, exist := filesByTarget[target]; exist {
			if existFile == file.Path {
				continue
			}
			return errN("scan resources error",
				reason("target duplicated"),
				kv("target", target),
				kv("file", file.Path),
				kv("existFile", existFile),
			)
		}
		filesByTarget[target] = file.Path

		switch file.Type {
		case utils.FileTypeTemplate:
			templateItem := &projectResourceTemplateItemEntity{
				File:   file.Path,
				Target: target,
			}
			e.TemplateItems = append(e.TemplateItems, templateItem)
		case utils.FileTypePlain:
			plainItem := &projectResourcePlainItemEntity{
				File:   file.Path,
				Target: target,
			}
			e.PlainItems = append(e.PlainItems, plainItem)
		default:
			impossible()
		}
	}
	return nil
}

func (e *projectResourceEntity) loadConfigFiles() (contents []*projectResourceConfigItemContentEntity, err error) {
	for i := 0; i < len(e.ConfigItems); i++ {
		config := e.ConfigItems[i]
		err = config.load()
		if err != nil {
			return nil, err
		}
		contents = append(contents, config.content)
	}
	return contents, nil
}

func (e *projectResourceEntity) makeTargetFiles(evaluator *Evaluator, outputPath string, useHardLink bool) (targetNames []string, err error) {
	for i := 0; i < len(e.PlainItems); i++ {
		startTime := time.Now()
		item := e.PlainItems[i]
		targetFile := filepath.Join(outputPath, item.Target)
		e.context.logger.InfoDesc("make script sources start",
			kv("sourceType", utils.FileTypePlain),
			kv("sourceFile", item.File),
			kv("targetFile", targetFile),
		)
		if useHardLink {
			err = utils.LinkOrCopyFile(item.File, targetFile)
			if err != nil {
				return nil, errW(err, "make script sources error",
					reason("link or copy file error"),
					kv("sourceType", utils.FileTypePlain),
					kv("sourceFile", item.File),
					kv("targetFile", targetFile),
				)
			}
		} else {
			err = utils.CopyFile(item.File, targetFile)
			if err != nil {
				return nil, errW(err, "make script sources error",
					reason("copy file error"),
					kv("sourceType", utils.FileTypePlain),
					kv("sourceFile", item.File),
					kv("targetFile", targetFile),
				)
			}
		}
		targetNames = append(targetNames, strings.ReplaceAll(item.Target, "\\", "/"))
		e.context.logger.InfoDesc("make script sources finish",
			kv("elapsed", time.Since(startTime)),
		)
	}
	var templateLibFiles []string
	for i := 0; i < len(e.TemplateLibItems); i++ {
		templateLibFiles = append(templateLibFiles, e.TemplateLibItems[i].File)
	}
	for i := 0; i < len(e.TemplateItems); i++ {
		startTime := time.Now()
		item := e.TemplateItems[i]
		targetFile := filepath.Join(outputPath, item.Target)
		e.context.logger.InfoDesc("make script sources start",
			kv("sourceType", utils.FileTypeTemplate),
			kv("sourceFile", item.File),
			kv("targetFile", targetFile),
		)
		if err = evaluator.EvalFileTemplate(item.File, templateLibFiles, targetFile); err != nil {
			return nil, errW(err, "make script sources error",
				reason("make template error"),
				kv("sourceType", utils.FileTypeTemplate),
				kv("sourceFile", item.File),
				kv("targetFile", targetFile),
			)
		}
		targetNames = append(targetNames, strings.ReplaceAll(item.Target, "\\", "/"))
		e.context.logger.InfoDesc("make script sources finish",
			kv("elapsed", time.Since(startTime)),
		)
	}
	return targetNames, nil
}

func (e *projectResourceEntity) inspect() *ProjectResourceEntityInspection {
	var configItems []*ProjectResourceConfigItemEntityInspection
	for i := 0; i < len(e.ConfigItems); i++ {
		configItems = append(configItems, e.ConfigItems[i].inspect())
	}
	var templateItems []*ProjectResourceTemplateItemEntityInspection
	for i := 0; i < len(e.TemplateItems); i++ {
		templateItems = append(templateItems, e.TemplateItems[i].inspect())
	}
	var templateLibItems []*ProjectResourceTemplateLibItemEntityInspection
	for i := 0; i < len(e.TemplateLibItems); i++ {
		templateLibItems = append(templateLibItems, e.TemplateLibItems[i].inspect())
	}
	var plainItems []*ProjectResourcePlainItemEntityInspection
	for i := 0; i < len(e.PlainItems); i++ {
		plainItems = append(plainItems, e.PlainItems[i].inspect())
	}
	return newProjectResourceEntityInspection(configItems, templateItems, templateLibItems, plainItems)
}

// endregion

// region projectResourceConfigItemEntity

type projectResourceConfigItemEntity struct {
	File    string
	Format  projectResourceConfigFormat
	content *projectResourceConfigItemContentEntity
}

func (e *projectResourceConfigItemEntity) load() error {
	if e.content == nil {
		if content, err := newProjectResourceConfigItemContentEntity(e.File, e.Format); err != nil {
			return err
		} else {
			e.content = content
		}
	}
	return nil
}

func (e *projectResourceConfigItemEntity) inspect() *ProjectResourceConfigItemEntityInspection {
	return newProjectResourceConfigItemEntityInspection(e.File, string(e.Format))
}

// endregion

// region projectResourceConfigItemContentEntity

type projectResourceConfigItemContentEntity struct {
	Order   int64
	Merges  map[string]projectResourceConfigMergeMode
	Configs map[string]any
	file    string
}

func newProjectResourceConfigItemContentEntity(file string, format projectResourceConfigFormat) (*projectResourceConfigItemContentEntity, error) {
	content := &projectResourceConfigItemContentEntity{
		Merges: map[string]projectResourceConfigMergeMode{},
		file:   file,
	}
	if _, err := utils.DeserializeFromFile(file, format, content); err != nil {
		return nil, errW(err, "load config sources error",
			reason("deserialize error"),
			kv("file", file),
			kv("format", format),
		)
	}
	for k, v := range content.Merges {
		switch v {
		case utils.MapMergeModeReplace:
		case utils.MapMergeModeInsert:
		default:
			return nil, errN("load config sources error",
				reason("merge mode invalid"),
				kv("file", file),
				kv("field", fmt.Sprintf("merges[%s]", k)),
				kv("value", v),
			)
		}
	}
	return content, nil
}

func (e *projectResourceConfigItemContentEntity) merge(configs map[string]any, configsTraces map[string]any) error {
	if _, _, err := utils.MapMerge(configs, e.Configs, e.Merges, e.file, configsTraces); err != nil {
		return errW(err, "merge config content error",
			kv("file", e.file),
		)
	}
	return nil
}

// endregion

// region projectResourceTemplateItemEntity

type projectResourceTemplateItemEntity struct {
	File   string
	Target string
}

func (e *projectResourceTemplateItemEntity) inspect() *ProjectResourceTemplateItemEntityInspection {
	return newProjectResourceTemplateItemEntityInspection(e.File, e.Target)
}

// endregion

// region projectResourceTemplateLibItemEntity

type projectResourceTemplateLibItemEntity struct {
	File string
}

func (e *projectResourceTemplateLibItemEntity) inspect() *ProjectResourceTemplateLibItemEntityInspection {
	return newProjectResourceTemplateLibItemEntityInspection(e.File)
}

// endregion

// region projectResourcePlainItemEntity

type projectResourcePlainItemEntity struct {
	File   string
	Target string
}

func (e *projectResourcePlainItemEntity) inspect() *ProjectResourcePlainItemEntityInspection {
	return newProjectResourcePlainItemEntityInspection(e.File, e.Target)
}

// endregion
