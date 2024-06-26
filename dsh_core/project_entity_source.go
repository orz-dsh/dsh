package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// region projectSourceEntity

type projectSourceEntity struct {
	context           *appContext
	ProjectName       string
	ConfigFiles       []*projectSourceConfigFileEntity
	TemplateFiles     []*projectSourceTargetFileEntity
	TemplateLibFiles  []*projectSourceTargetFileEntity
	PlainFiles        []*projectSourceTargetFileEntity
	configFilesDict   map[string]bool
	targetFilesByName map[string]*projectSourceTargetFileEntity
}

func newProjectSourceEntity(context *appContext, setting *projectSetting, option *projectOptionEntity) (*projectSourceEntity, error) {
	source := &projectSourceEntity{
		context:           context,
		ProjectName:       setting.Name,
		configFilesDict:   map[string]bool{},
		targetFilesByName: map[string]*projectSourceTargetFileEntity{},
	}
	for i := 0; i < len(setting.Resource.Items); i++ {
		sourceSetting := setting.Resource.Items[i]
		matched, err := option.evaluator.EvalBoolExpr(sourceSetting.match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if err = source.scan(filepath.Join(setting.Dir, sourceSetting.Dir), sourceSetting.Includes); err != nil {
			return nil, err
		}
	}
	return source, nil
}

func (e *projectSourceEntity) scan(dir string, includeFiles []string) error {
	files, err := dsh_utils.ScanFiles(dir, includeFiles, []dsh_utils.FileType{
		dsh_utils.FileTypeConfigYaml,
		dsh_utils.FileTypeConfigToml,
		dsh_utils.FileTypeConfigJson,
		dsh_utils.FileTypeTemplate,
		dsh_utils.FileTypeTemplateLib,
		dsh_utils.FileTypePlain,
	})
	if err != nil {
		return err
	}
	for i := 0; i < len(files); i++ {
		file := files[i]
		switch file.Type {
		case dsh_utils.FileTypeConfigYaml, dsh_utils.FileTypeConfigToml, dsh_utils.FileTypeConfigJson:
			if !e.configFilesDict[file.Path] {
				configFile := &projectSourceConfigFileEntity{
					Path:   file.Path,
					Format: dsh_utils.GetSerializationFormat(file.Type),
				}
				e.ConfigFiles = append(e.ConfigFiles, configFile)
				e.configFilesDict[file.Path] = true
			}
			continue
		}
		targetFile := &projectSourceTargetFileEntity{
			Path: file.Path,
			Name: file.RelPath,
		}
		if file.Type == dsh_utils.FileTypeTemplate {
			targetFile.Name = targetFile.Name[:len(targetFile.Name)-len(".dtpl")]
		}
		if existTargetFile, exist := e.targetFilesByName[targetFile.Name]; exist {
			if existTargetFile.Path == targetFile.Path {
				continue
			}
			return errN("scan script sources error",
				reason("target file name duplicated"),
				kv("targetFile", targetFile),
				kv("existTargetFile", existTargetFile),
			)
		}
		switch file.Type {
		case dsh_utils.FileTypeTemplate:
			e.TemplateFiles = append(e.TemplateFiles, targetFile)
		case dsh_utils.FileTypeTemplateLib:
			e.TemplateLibFiles = append(e.TemplateLibFiles, targetFile)
		case dsh_utils.FileTypePlain:
			e.PlainFiles = append(e.PlainFiles, targetFile)
		default:
			impossible()
		}
		e.targetFilesByName[targetFile.Name] = targetFile
	}
	return nil
}

func (e *projectSourceEntity) loadConfigFiles() (contents []*projectSourceConfigContentEntity, err error) {
	for i := 0; i < len(e.ConfigFiles); i++ {
		config := e.ConfigFiles[i]
		err = config.load()
		if err != nil {
			return nil, err
		}
		contents = append(contents, config.content)
	}
	return contents, nil
}

func (e *projectSourceEntity) makeTargetFiles(evaluator *Evaluator, outputPath string, useHardLink bool) (targetNames []string, err error) {
	for i := 0; i < len(e.PlainFiles); i++ {
		startTime := time.Now()
		source := e.PlainFiles[i]
		target := filepath.Join(e.ProjectName, source.Name)
		targetPath := filepath.Join(outputPath, target)
		e.context.logger.InfoDesc("make script sources start",
			kv("sourceType", dsh_utils.FileTypePlain),
			kv("sourcePath", source.Path),
			kv("targetPath", targetPath),
		)
		if useHardLink {
			err = dsh_utils.LinkOrCopyFile(source.Path, targetPath)
			if err != nil {
				return nil, errW(err, "make script sources error",
					reason("link or copy file error"),
					kv("sourceType", dsh_utils.FileTypePlain),
					kv("sourcePath", source.Path),
					kv("targetPath", targetPath),
				)
			}
		} else {
			err = dsh_utils.CopyFile(source.Path, targetPath)
			if err != nil {
				return nil, errW(err, "make script sources error",
					reason("copy file error"),
					kv("sourceType", dsh_utils.FileTypePlain),
					kv("sourcePath", source.Path),
					kv("targetPath", targetPath),
				)
			}
		}
		targetNames = append(targetNames, strings.ReplaceAll(target, "\\", "/"))
		e.context.logger.InfoDesc("make script sources finish",
			kv("elapsed", time.Since(startTime)),
		)
	}
	var templateLibSourcePaths []string
	for i := 0; i < len(e.TemplateLibFiles); i++ {
		templateLibSourcePaths = append(templateLibSourcePaths, e.TemplateLibFiles[i].Path)
	}
	for i := 0; i < len(e.TemplateFiles); i++ {
		startTime := time.Now()
		source := e.TemplateFiles[i]
		target := filepath.Join(e.ProjectName, source.Name)
		targetPath := filepath.Join(outputPath, target)
		e.context.logger.InfoDesc("make script sources start",
			kv("sourceType", dsh_utils.FileTypeTemplate),
			kv("sourcePath", source.Path),
			kv("targetPath", targetPath),
		)
		if err = evaluator.EvalFileTemplate(source.Path, templateLibSourcePaths, targetPath); err != nil {
			return nil, errW(err, "make script sources error",
				reason("make template error"),
				kv("sourceType", dsh_utils.FileTypeTemplate),
				kv("sourcePath", source.Path),
				kv("targetPath", targetPath),
			)
		}
		targetNames = append(targetNames, strings.ReplaceAll(target, "\\", "/"))
		e.context.logger.InfoDesc("make script sources finish",
			kv("elapsed", time.Since(startTime)),
		)
	}
	return targetNames, nil
}

func (e *projectSourceEntity) inspect() *ProjectSourceEntityInspection {
	var configFiles []*ProjectSourceConfigFileEntityInspection
	for i := 0; i < len(e.ConfigFiles); i++ {
		configFiles = append(configFiles, e.ConfigFiles[i].inspect())
	}
	var templateFiles []*ProjectSourceTargetFileEntityInspection
	for i := 0; i < len(e.TemplateFiles); i++ {
		templateFiles = append(templateFiles, e.TemplateFiles[i].inspect())
	}
	var templateLibFiles []*ProjectSourceTargetFileEntityInspection
	for i := 0; i < len(e.TemplateLibFiles); i++ {
		templateLibFiles = append(templateLibFiles, e.TemplateLibFiles[i].inspect())
	}
	var plainFiles []*ProjectSourceTargetFileEntityInspection
	for i := 0; i < len(e.PlainFiles); i++ {
		plainFiles = append(plainFiles, e.PlainFiles[i].inspect())
	}
	return newProjectSourceEntityInspection(configFiles, templateFiles, templateLibFiles, plainFiles)
}

// endregion

// region projectSourceConfigFileEntity

type projectSourceConfigFileEntity struct {
	Path    string
	Format  projectSourceConfigFormat
	content *projectSourceConfigContentEntity
}

type projectSourceConfigFormat = dsh_utils.SerializationFormat

func (e *projectSourceConfigFileEntity) load() error {
	if e.content == nil {
		if content, err := newProjectSourceConfigContentEntity(e.Path, e.Format); err != nil {
			return err
		} else {
			e.content = content
		}
	}
	return nil
}

func (e *projectSourceConfigFileEntity) inspect() *ProjectSourceConfigFileEntityInspection {
	return newProjectSourceConfigFileEntityInspection(e.Path, string(e.Format))
}

// endregion

// region projectSourceConfigContentEntity

type projectSourceConfigContentEntity struct {
	Order   int64
	Merges  map[string]projectSourceConfigContentMergeMode
	Configs map[string]any
	path    string
}

type projectSourceConfigContentMergeMode = dsh_utils.MapMergeMode

func newProjectSourceConfigContentEntity(sourcePath string, sourceFormat projectSourceConfigFormat) (*projectSourceConfigContentEntity, error) {
	content := &projectSourceConfigContentEntity{
		Merges: map[string]projectSourceConfigContentMergeMode{},
		path:   sourcePath,
	}
	if _, err := dsh_utils.DeserializeFromFile(sourcePath, sourceFormat, content); err != nil {
		return nil, errW(err, "load config sources error",
			reason("deserialize error"),
			kv("sourcePath", sourceFormat),
			kv("sourceFormat", sourcePath),
		)
	}
	for k, v := range content.Merges {
		switch v {
		case dsh_utils.MapMergeModeReplace:
		case dsh_utils.MapMergeModeInsert:
		default:
			return nil, errN("load config sources error",
				reason("merge mode invalid"),
				kv("file", sourcePath),
				kv("field", fmt.Sprintf("merges[%s]", k)),
				kv("value", v),
			)
		}
	}
	return content, nil
}

func (e *projectSourceConfigContentEntity) merge(configs map[string]any, configsTraces map[string]any) error {
	if _, _, err := dsh_utils.MapMerge(configs, e.Configs, e.Merges, e.path, configsTraces); err != nil {
		return errW(err, "merge config content error",
			kv("sourcePath", e.path),
		)
	}
	return nil
}

// endregion

// region projectSourceTargetFileEntity

type projectSourceTargetFileEntity struct {
	Path string
	Name string
}

func (e *projectSourceTargetFileEntity) inspect() *ProjectSourceTargetFileEntityInspection {
	return newProjectSourceTargetFileEntityInspection(e.Path, e.Name)
}

// endregion

// region ProjectSourceEntityInspection

type ProjectSourceEntityInspection struct {
	ConfigFiles      []*ProjectSourceConfigFileEntityInspection `yaml:"configFiles" toml:"configFiles" json:"configFiles"`
	TemplateFiles    []*ProjectSourceTargetFileEntityInspection `yaml:"templateFiles" toml:"templateFiles" json:"templateFiles"`
	TemplateLibFiles []*ProjectSourceTargetFileEntityInspection `yaml:"templateLibFiles" toml:"templateLibFiles" json:"templateLibFiles"`
	PlainFiles       []*ProjectSourceTargetFileEntityInspection `yaml:"plainFiles" toml:"plainFiles" json:"plainFiles"`
}

func newProjectSourceEntityInspection(configFiles []*ProjectSourceConfigFileEntityInspection, templateFiles []*ProjectSourceTargetFileEntityInspection, templateLibFiles []*ProjectSourceTargetFileEntityInspection, plainFiles []*ProjectSourceTargetFileEntityInspection) *ProjectSourceEntityInspection {
	return &ProjectSourceEntityInspection{
		ConfigFiles:      configFiles,
		TemplateFiles:    templateFiles,
		TemplateLibFiles: templateLibFiles,
		PlainFiles:       plainFiles,
	}
}

// endregion

// region ProjectSourceConfigFileEntityInspection

type ProjectSourceConfigFileEntityInspection struct {
	Path   string `yaml:"path" toml:"path" json:"path"`
	Format string `yaml:"format" toml:"format" json:"format"`
}

func newProjectSourceConfigFileEntityInspection(path string, format string) *ProjectSourceConfigFileEntityInspection {
	return &ProjectSourceConfigFileEntityInspection{
		Path:   path,
		Format: format,
	}
}

// endregion

// region ProjectSourceTargetFileEntityInspection

type ProjectSourceTargetFileEntityInspection struct {
	Path string `yaml:"path" toml:"path" json:"path"`
	Name string `yaml:"name" toml:"name" json:"name"`
}

func newProjectSourceTargetFileEntityInspection(path string, name string) *ProjectSourceTargetFileEntityInspection {
	return &ProjectSourceTargetFileEntityInspection{
		Path: path,
		Name: name,
	}
}

// endregion
