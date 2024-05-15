package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"reflect"
)

type projectConfig struct {
	sourceContainer *projectConfigSourceContainer
	importContainer *projectImportContainer
}

type projectConfigSource struct {
	sourcePath string
	sourceName string
	sourceType projectConfigSourceType
	content    *projectConfigSourceContent
}

type projectConfigSourceContainer struct {
	context       *appContext
	sources       []*projectConfigSource
	sourcesByName map[string]*projectConfigSource
}

type projectConfigSourceContent struct {
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

func loadProjectConfig(context *appContext, manifest *projectManifest) (pc *projectConfig, err error) {
	sc, err := loadProjectConfigSourceContainer(context, manifest)
	if err != nil {
		return nil, err
	}
	ic, err := loadProjectImportContainer(context, manifest, projectImportScopeConfig)
	if err != nil {
		return nil, err
	}
	pc = &projectConfig{
		sourceContainer: sc,
		importContainer: ic,
	}
	return pc, nil
}

func loadProjectConfigSourceContainer(context *appContext, manifest *projectManifest) (sc *projectConfigSourceContainer, err error) {
	sc = &projectConfigSourceContainer{
		context:       context,
		sourcesByName: make(map[string]*projectConfigSource),
	}
	for i := 0; i < len(manifest.Config.Sources); i++ {
		src := manifest.Config.Sources[i]
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
			if err = sc.scanSources(filepath.Join(manifest.projectPath, src.Dir), src.Files); err != nil {
				return nil, err
			}
		}
	}
	return sc, nil
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
			// impossible
			panic(desc("config source type unsupported",
				kv("filePath", filePath),
				kv("fileType", fileType),
			))
		}
		source := &projectConfigSource{
			sourcePath: filepath.Join(sourceDir, filePath),
			sourceName: dsh_utils.RemoveFileExt(filePath),
			sourceType: sourceType,
		}
		if existSource, exist := c.sourcesByName[source.sourcePath]; exist {
			if existSource.sourcePath == source.sourcePath {
				continue
			}
			return errN("scan config sources error",
				reason("source name duplicated"),
				kv("sourceName", source.sourceName),
				kv("sourcePath1", source.sourcePath),
				kv("sourcePath2", existSource.sourcePath),
			)
		}
		c.sources = append(c.sources, source)
		c.sourcesByName[source.sourceName] = source
	}
	return nil
}

func (c *projectConfigSourceContainer) loadSources() (err error) {
	for i := 0; i < len(c.sources); i++ {
		source := c.sources[i]
		if source.content == nil {
			content := &projectConfigSourceContent{}
			switch source.sourceType {
			case projectConfigSourceTypeYaml:
				if err = dsh_utils.ReadYamlFile(source.sourcePath, content); err != nil {
					return errW(err, "load config sources error",
						reason("read yaml file error"),
						kv("sourcePath", source.sourcePath),
					)
				}
			case projectConfigSourceTypeToml:
				if err = dsh_utils.ReadTomlFile(source.sourcePath, content); err != nil {
					return errW(err, "load config sources error",
						reason("read toml file error"),
						kv("sourcePath", source.sourcePath),
					)
				}
			case projectConfigSourceTypeJson:
				if err = dsh_utils.ReadJsonFile(source.sourcePath, content); err != nil {
					return errW(err, "load config sources error",
						reason("read json file error"),
						kv("sourcePath", source.sourcePath),
					)
				}
			default:
				// impossible
				panic(desc("config source type unsupported",
					kv("sourcePath", source.sourcePath),
					kv("sourceType", source.sourceType),
				))
			}
			for k, v := range content.Merges {
				switch v {
				case projectConfigMergeTypeReplace:
				case projectConfigMergeTypeInsert:
				default:
					return errN("load config sources error",
						reason("merge type invalid"),
						kv("sourcePath", source.sourcePath),
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
		target = make(map[string]any)
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
							kv("sourcePath", s.sourcePath),
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
					kv("sourcePath", s.sourcePath),
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
							kv("sourcePath", s.sourcePath),
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
					kv("sourcePath", s.sourcePath),
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
