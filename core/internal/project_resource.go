package internal

import (
	"fmt"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"path/filepath"
	"strings"
	"time"
)

// region base

type projectResourceConfigFormat = SerializationFormat

type projectResourceConfigMergeMode = MapMergeMode

// endregion

// region ProjectResource

type ProjectResource struct {
	context          *ApplicationCore
	ConfigItems      []*ProjectResourceConfigItem
	TemplateItems    []*ProjectResourceTemplateItem
	TemplateLibItems []*ProjectResourceTemplateLibItem
	PlainItems       []*ProjectResourcePlainItem
}

func NewProjectResource(context *ApplicationCore, setting *ProjectSetting, option *ProjectOption) (*ProjectResource, error) {
	resource := &ProjectResource{context: context}
	configItemsDict := map[string]bool{}
	templateLibItemsDict := map[string]bool{}
	filesByTarget := map[string]string{}
	for i := 0; i < len(setting.Resource.Items); i++ {
		item := setting.Resource.Items[i]
		matched, err := option.evaluator.EvalBoolExpr(item.MatchObj)
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

func (e *ProjectResource) scan(projectName, dir string, includes, excludes []string, configItemsDict, templateLibItemsDict map[string]bool, filesByTarget map[string]string) error {
	files, err := ScanFiles(dir, includes, excludes, []FileType{
		FileTypeConfigYaml,
		FileTypeConfigToml,
		FileTypeConfigJson,
		FileTypeTemplate,
		FileTypeTemplateLib,
		FileTypePlain,
	})
	if err != nil {
		return err
	}
	for i := 0; i < len(files); i++ {
		file := files[i]
		switch file.Type {
		case FileTypeConfigYaml, FileTypeConfigToml, FileTypeConfigJson:
			if !configItemsDict[file.Path] {
				configItem := &ProjectResourceConfigItem{
					File:   file.Path,
					Format: GetSerializationFormat(file.Type),
				}
				e.ConfigItems = append(e.ConfigItems, configItem)
				configItemsDict[file.Path] = true
			}
			continue
		case FileTypeTemplateLib:
			if !templateLibItemsDict[file.Path] {
				templateLibItem := &ProjectResourceTemplateLibItem{
					File: file.Path,
				}
				e.TemplateLibItems = append(e.TemplateLibItems, templateLibItem)
				templateLibItemsDict[file.Path] = true
			}
			continue
		}
		target := filepath.Join(projectName, file.RelPath)
		if file.Type == FileTypeTemplate {
			target = target[:len(target)-len(".dtpl")]
		}
		if existFile, exist := filesByTarget[target]; exist {
			if existFile == file.Path {
				continue
			}
			return ErrN("scan resources error",
				Reason("target duplicated"),
				KV("target", target),
				KV("file", file.Path),
				KV("existFile", existFile),
			)
		}
		filesByTarget[target] = file.Path

		switch file.Type {
		case FileTypeTemplate:
			templateItem := &ProjectResourceTemplateItem{
				File:   file.Path,
				Target: target,
			}
			e.TemplateItems = append(e.TemplateItems, templateItem)
		case FileTypePlain:
			plainItem := &ProjectResourcePlainItem{
				File:   file.Path,
				Target: target,
			}
			e.PlainItems = append(e.PlainItems, plainItem)
		default:
			Impossible()
		}
	}
	return nil
}

func (e *ProjectResource) loadConfigFiles() (contents []*ProjectResourceConfigItemContent, err error) {
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

func (e *ProjectResource) makeTargetFiles(evaluator *Evaluator, outputPath string, useHardLink bool) (targetNames []string, err error) {
	for i := 0; i < len(e.PlainItems); i++ {
		startTime := time.Now()
		item := e.PlainItems[i]
		targetFile := filepath.Join(outputPath, item.Target)
		e.context.Logger.InfoDesc("make script sources start",
			KV("sourceType", FileTypePlain),
			KV("sourceFile", item.File),
			KV("targetFile", targetFile),
		)
		if useHardLink {
			err = LinkOrCopyFile(item.File, targetFile)
			if err != nil {
				return nil, ErrW(err, "make script sources error",
					Reason("link or copy file error"),
					KV("sourceType", FileTypePlain),
					KV("sourceFile", item.File),
					KV("targetFile", targetFile),
				)
			}
		} else {
			err = CopyFile(item.File, targetFile)
			if err != nil {
				return nil, ErrW(err, "make script sources error",
					Reason("copy file error"),
					KV("sourceType", FileTypePlain),
					KV("sourceFile", item.File),
					KV("targetFile", targetFile),
				)
			}
		}
		targetNames = append(targetNames, strings.ReplaceAll(item.Target, "\\", "/"))
		e.context.Logger.InfoDesc("make script sources finish",
			KV("elapsed", time.Since(startTime)),
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
		e.context.Logger.InfoDesc("make script sources start",
			KV("sourceType", FileTypeTemplate),
			KV("sourceFile", item.File),
			KV("targetFile", targetFile),
		)
		if err = evaluator.EvalFileTemplate(item.File, templateLibFiles, targetFile); err != nil {
			return nil, ErrW(err, "make script sources error",
				Reason("make template error"),
				KV("sourceType", FileTypeTemplate),
				KV("sourceFile", item.File),
				KV("targetFile", targetFile),
			)
		}
		targetNames = append(targetNames, strings.ReplaceAll(item.Target, "\\", "/"))
		e.context.Logger.InfoDesc("make script sources finish",
			KV("elapsed", time.Since(startTime)),
		)
	}
	return targetNames, nil
}

func (e *ProjectResource) inspect() *ProjectResourceInspection {
	var configItems []*ProjectResourceConfigItemInspection
	for i := 0; i < len(e.ConfigItems); i++ {
		configItems = append(configItems, e.ConfigItems[i].inspect())
	}
	var templateItems []*ProjectResourceTemplateItemInspection
	for i := 0; i < len(e.TemplateItems); i++ {
		templateItems = append(templateItems, e.TemplateItems[i].inspect())
	}
	var templateLibItems []*ProjectResourceTemplateLibItemInspection
	for i := 0; i < len(e.TemplateLibItems); i++ {
		templateLibItems = append(templateLibItems, e.TemplateLibItems[i].inspect())
	}
	var plainItems []*ProjectResourcePlainItemInspection
	for i := 0; i < len(e.PlainItems); i++ {
		plainItems = append(plainItems, e.PlainItems[i].inspect())
	}
	return NewProjectResourceInspection(configItems, templateItems, templateLibItems, plainItems)
}

// endregion

// region ProjectResourceConfigItem

type ProjectResourceConfigItem struct {
	File    string
	Format  projectResourceConfigFormat
	content *ProjectResourceConfigItemContent
}

func (e *ProjectResourceConfigItem) load() error {
	if e.content == nil {
		if content, err := newProjectResourceConfigItemContentEntity(e.File, e.Format); err != nil {
			return err
		} else {
			e.content = content
		}
	}
	return nil
}

func (e *ProjectResourceConfigItem) inspect() *ProjectResourceConfigItemInspection {
	return NewProjectResourceConfigItemInspection(e.File, string(e.Format))
}

// endregion

// region ProjectResourceConfigItemContent

type ProjectResourceConfigItemContent struct {
	Order   int64
	Merges  map[string]projectResourceConfigMergeMode
	Configs map[string]any
	file    string
}

func newProjectResourceConfigItemContentEntity(file string, format projectResourceConfigFormat) (*ProjectResourceConfigItemContent, error) {
	content := &ProjectResourceConfigItemContent{
		Merges: map[string]projectResourceConfigMergeMode{},
		file:   file,
	}
	if _, err := DeserializeFromFile(file, format, content); err != nil {
		return nil, ErrW(err, "load config sources error",
			Reason("deserialize error"),
			KV("file", file),
			KV("format", format),
		)
	}
	for k, v := range content.Merges {
		switch v {
		case MapMergeModeReplace:
		case MapMergeModeInsert:
		default:
			return nil, ErrN("load config sources error",
				Reason("merge mode invalid"),
				KV("file", file),
				KV("field", fmt.Sprintf("merges[%s]", k)),
				KV("value", v),
			)
		}
	}
	return content, nil
}

func (e *ProjectResourceConfigItemContent) merge(configs map[string]any, configsTraces map[string]any) error {
	if _, _, err := MapMerge(configs, e.Configs, e.Merges, e.file, configsTraces); err != nil {
		return ErrW(err, "merge config content error",
			KV("file", e.file),
		)
	}
	return nil
}

// endregion

// region ProjectResourceTemplateItem

type ProjectResourceTemplateItem struct {
	File   string
	Target string
}

func (e *ProjectResourceTemplateItem) inspect() *ProjectResourceTemplateItemInspection {
	return NewProjectResourceTemplateItemInspection(e.File, e.Target)
}

// endregion

// region ProjectResourceTemplateLibItem

type ProjectResourceTemplateLibItem struct {
	File string
}

func (e *ProjectResourceTemplateLibItem) inspect() *ProjectResourceTemplateLibItemInspection {
	return NewProjectResourceTemplateLibItemInspection(e.File)
}

// endregion

// region ProjectResourcePlainItem

type ProjectResourcePlainItem struct {
	File   string
	Target string
}

func (e *ProjectResourcePlainItem) inspect() *ProjectResourcePlainItemInspection {
	return NewProjectResourcePlainItemInspection(e.File, e.Target)
}

// endregion
