package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"reflect"
)

// region config

type projectConfig struct {
	SourceContainer *projectConfigSourceContainer
	ImportContainer *projectImportContainer
}

func makeProjectConfig(context *appContext, entity *projectEntity, option *projectOption) (config *projectConfig, err error) {
	sc, err := makeProjectConfigSourceContainer(context, entity, option)
	if err != nil {
		return nil, err
	}
	ic, err := makeProjectImportContainer(context, entity, option, projectImportScopeConfig)
	if err != nil {
		return nil, err
	}
	config = &projectConfig{
		SourceContainer: sc,
		ImportContainer: ic,
	}
	return config, nil
}

// endregion

// region source

type projectConfigSource struct {
	SourcePath string
	SourceName string
	SourceType projectConfigSourceType
	content    *projectConfigContent
}

type projectConfigContent struct {
	Order   int64
	Merges  map[string]string
	Configs map[string]any
}

type projectConfigSourceType string

const (
	projectConfigSourceTypeYaml projectConfigSourceType = "yaml"
	projectConfigSourceTypeToml projectConfigSourceType = "toml"
	projectConfigSourceTypeJson projectConfigSourceType = "json"
)

const (
	projectConfigMergeKeyRoot     = "$root"
	projectConfigMergeTypeReplace = "replace"
	projectConfigMergeTypeInsert  = "insert"
)

func (s *projectConfigSource) mergeConfigs(configs map[string]any) error {
	content := s.content
	if content.Merges[projectConfigMergeKeyRoot] == projectConfigMergeTypeReplace {
		clear(configs)
		if _, err := s.merge(configs, content.Configs, ""); err != nil {
			return err
		}
	} else if _, err := s.merge(configs, content.Configs, ""); err != nil {
		return err
	}
	return nil
}

func (s *projectConfigSource) merge(target map[string]any, source map[string]any, key string) (_ map[string]any, err error) {
	if target == nil {
		target = map[string]any{}
	}
	for k, v := range source {
		switch v.(type) {
		case map[string]any:
			sourceKey := k
			if key != "" {
				sourceKey = key + "." + k
			}
			sourceMap := v.(map[string]any)
			targetValue := target[k]
			if targetValue == nil {
				if target[k], err = s.merge(nil, sourceMap, sourceKey); err != nil {
					return nil, err
				}
			} else if targetMap, ok := targetValue.(map[string]any); ok {
				if merge, exist := s.content.Merges[sourceKey]; exist {
					if merge == projectConfigMergeTypeReplace {
						if target[k], err = s.merge(nil, sourceMap, sourceKey); err != nil {
							return nil, err
						}
					} else {
						return nil, errN("merge configs error",
							reason("merge type invalid"),
							kv("sourcePath", s.SourcePath),
							kv("sourceKey", sourceKey),
							kv("mergeType", merge),
							kv("supportType", []string{
								projectConfigMergeTypeReplace,
							}),
						)
					}
				} else {
					if target[k], err = s.merge(targetMap, sourceMap, sourceKey); err != nil {
						return nil, err
					}
				}
			} else {
				return nil, errN("merge configs error",
					reason("source type not match target type"),
					kv("sourcePath", s.SourcePath),
					kv("sourceKey", sourceKey),
					kv("sourceType", reflect.TypeOf(sourceMap)),
					kv("targetType", reflect.TypeOf(targetValue)),
				)
			}
		case []any:
			sourceKey := k
			if key != "" {
				sourceKey = key + "." + k
			}
			sourceList := v.([]any)
			targetValue := target[k]
			if targetValue == nil {
				target[k] = sourceList
			} else if targetList, ok := targetValue.([]any); ok {
				if mergeKey, exist := s.content.Merges[sourceKey]; exist {
					if mergeKey == projectConfigMergeTypeReplace {
						target[k] = sourceList
					} else if mergeKey == projectConfigMergeTypeInsert {
						target[k] = append(sourceList, targetList...)
					} else {
						return nil, errN("merge configs error",
							reason("merge type invalid"),
							kv("sourcePath", s.SourcePath),
							kv("sourceKey", sourceKey),
							kv("mergeType", mergeKey),
							kv("supportType", []string{
								projectConfigMergeTypeReplace,
								projectConfigMergeTypeInsert,
							}),
						)
					}
				} else {
					target[k] = append(targetList, sourceList...)
				}
			} else {
				return nil, errN("merge configs error",
					reason("source type not match target type"),
					kv("sourcePath", s.SourcePath),
					kv("sourceKey", sourceKey),
					kv("sourceType", reflect.TypeOf(sourceList)),
					kv("targetType", reflect.TypeOf(targetValue)),
				)
			}
			break
		default:
			target[k] = v
		}
	}
	return target, nil
}

// endregion

// region container

type projectConfigSourceContainer struct {
	context         *appContext
	Sources         []*projectConfigSource
	sourcePathsDict map[string]bool
}

func makeProjectConfigSourceContainer(context *appContext, entity *projectEntity, option *projectOption) (container *projectConfigSourceContainer, err error) {
	container = &projectConfigSourceContainer{
		context:         context,
		sourcePathsDict: map[string]bool{},
	}
	sources := entity.ConfigSources
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

func (c *projectConfigSourceContainer) scanSources(sourceDir string, includeFiles []string) error {
	filePaths, fileTypes, err := dsh_utils.ScanFiles(sourceDir, includeFiles, []dsh_utils.FileType{
		dsh_utils.FileTypeYaml,
		dsh_utils.FileTypeToml,
		dsh_utils.FileTypeJson,
	})
	if err != nil {
		return err
	}
	for i := 0; i < len(filePaths); i++ {
		filePath := filePaths[i]
		fileType := fileTypes[i]
		var sourceType projectConfigSourceType
		switch fileType {
		case dsh_utils.FileTypeYaml:
			sourceType = projectConfigSourceTypeYaml
		case dsh_utils.FileTypeToml:
			sourceType = projectConfigSourceTypeToml
		case dsh_utils.FileTypeJson:
			sourceType = projectConfigSourceTypeJson
		default:
			impossible()
		}
		source := &projectConfigSource{
			SourcePath: filepath.Join(sourceDir, filePath),
			SourceName: filePath,
			SourceType: sourceType,
		}
		if c.sourcePathsDict[source.SourcePath] {
			continue
		}
		c.Sources = append(c.Sources, source)
		c.sourcePathsDict[source.SourcePath] = true
	}
	return nil
}

func (c *projectConfigSourceContainer) loadSources() (err error) {
	for i := 0; i < len(c.Sources); i++ {
		source := c.Sources[i]
		if source.content == nil {
			content := &projectConfigContent{}
			switch source.SourceType {
			case projectConfigSourceTypeYaml:
				if err = dsh_utils.ReadYamlFile(source.SourcePath, content); err != nil {
					return errW(err, "load config sources error",
						reason("read yaml file error"),
						kv("sourcePath", source.SourcePath),
					)
				}
			case projectConfigSourceTypeToml:
				if err = dsh_utils.ReadTomlFile(source.SourcePath, content); err != nil {
					return errW(err, "load config sources error",
						reason("read toml file error"),
						kv("sourcePath", source.SourcePath),
					)
				}
			case projectConfigSourceTypeJson:
				if err = dsh_utils.ReadJsonFile(source.SourcePath, content); err != nil {
					return errW(err, "load config sources error",
						reason("read json file error"),
						kv("sourcePath", source.SourcePath),
					)
				}
			default:
				impossible()
			}
			for k, v := range content.Merges {
				switch v {
				case projectConfigMergeTypeReplace:
				case projectConfigMergeTypeInsert:
				default:
					return errN("load config sources error",
						reason("merge type invalid"),
						kv("sourcePath", source.SourcePath),
						kv("field", fmt.Sprintf("merges[%s]", k)),
						kv("value", v),
					)
				}
			}
			source.content = content
		}
	}
	return nil
}

// endregion
