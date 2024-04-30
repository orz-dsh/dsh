package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
)

type projectInstanceConfig struct {
	sourceContainer *projectInstanceConfigSourceContainer
	importContainer *projectInstanceImportShallowContainer
}

type projectInstanceConfigSource struct {
	sourcePath string
	sourceName string
	sourceType projectInstanceConfigSourceType
	content    *projectInstanceConfigSourceContent
}

type projectInstanceConfigSourceContainer struct {
	context       *projectContext
	sourceNameMap map[string]*projectInstanceConfigSource
	sources       []*projectInstanceConfigSource
}

type projectInstanceConfigSourceContent struct {
	Order   int64
	Merges  map[string]string
	Configs map[string]any
}

type projectInstanceConfigSourceType string

const (
	projectInstanceConfigSourceTypeYaml projectInstanceConfigSourceType = "yaml"
	projectInstanceConfigSourceTypeToml projectInstanceConfigSourceType = "toml"
	projectInstanceConfigSourceTypeJson projectInstanceConfigSourceType = "json"
)

const (
	projectConfigMergeRoot        = "$root"
	projectConfigMergeTypeReplace = "replace"
	projectConfigMergeTypeInsert  = "insert"
)

func newProjectInstanceConfig(context *projectContext) *projectInstanceConfig {
	return &projectInstanceConfig{
		sourceContainer: &projectInstanceConfigSourceContainer{
			context:       context,
			sourceNameMap: make(map[string]*projectInstanceConfigSource),
		},
		importContainer: newProjectInstanceImportShallowContainer(context, projectInstanceImportScopeConfig),
	}
}

func (container *projectInstanceConfigSourceContainer) scanSources(sourceDir string, includeFiles []string) error {
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
		var sourceType projectInstanceConfigSourceType
		switch fileType {
		case dsh_utils.FileTypeYaml:
			sourceType = projectInstanceConfigSourceTypeYaml
		case dsh_utils.FileTypeToml:
			sourceType = projectInstanceConfigSourceTypeToml
		case dsh_utils.FileTypeJson:
			sourceType = projectInstanceConfigSourceTypeJson
		default:
			panic(fmt.Sprintf("unsupported config source type: filePath=%s, fileType=%s", filePath, fileType))
		}
		source := &projectInstanceConfigSource{
			sourcePath: filepath.Join(sourceDir, filePath),
			sourceName: dsh_utils.RemoveFileExt(filePath),
			sourceType: sourceType,
		}
		if existSource, exist := container.sourceNameMap[source.sourcePath]; exist {
			if existSource.sourcePath == source.sourcePath {
				continue
			}
			return dsh_utils.NewError("config source name is duplicated", map[string]any{
				"sourceName":  source.sourceName,
				"sourcePath1": source.sourcePath,
				"sourcePath2": existSource.sourcePath,
			})
		}
		container.sourceNameMap[source.sourceName] = source
		container.sources = append(container.sources, source)
	}
	return nil
}

func (container *projectInstanceConfigSourceContainer) loadSources() (err error) {
	for i := 0; i < len(container.sources); i++ {
		source := container.sources[i]
		if source.content == nil {
			content := &projectInstanceConfigSourceContent{}
			if source.sourceType == projectInstanceConfigSourceTypeYaml {
				if err = dsh_utils.ReadYamlFile(source.sourcePath, content); err != nil {
					return err
				}
			} else if source.sourceType == projectInstanceConfigSourceTypeToml {
				if err = dsh_utils.ReadTomlFile(source.sourcePath, content); err != nil {
					return err
				}
			} else if source.sourceType == projectInstanceConfigSourceTypeJson {
				if err = dsh_utils.ReadJsonFile(source.sourcePath, content); err != nil {
					return err
				}
			} else {
				panic(fmt.Sprintf("unsupported config source type: sourcePath=%s", source.sourcePath))
			}
			source.content = content
		}
	}
	return nil
}

func (content *projectInstanceConfigSourceContent) merge(target map[string]any) error {
	if content.Merges[projectConfigMergeRoot] == projectConfigMergeTypeReplace {
		clear(target)
		if _, err := mergeMap(target, content.Configs, content.Merges, ""); err != nil {
			return err
		}
	} else if _, err := mergeMap(target, content.Configs, content.Merges, ""); err != nil {
		return err
	}
	return nil
}

func mergeMap(target map[string]any, source map[string]any, merges map[string]string, key string) (_ map[string]any, err error) {
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
				if target[k], err = mergeMap(nil, sourceMap, merges, sourceKey); err != nil {
					return nil, err
				}
			} else if targetMap, ok := targetValue.(map[string]any); ok {
				if mergeKey, exist := merges[sourceKey]; exist {
					if mergeKey == projectConfigMergeTypeReplace {
						if target[k], err = mergeMap(nil, sourceMap, merges, sourceKey); err != nil {
							return nil, err
						}
					} else {
						if target[k], err = mergeMap(targetMap, sourceMap, merges, sourceKey); err != nil {
							return nil, err
						}
					}
				} else {
					if target[k], err = mergeMap(targetMap, sourceMap, merges, sourceKey); err != nil {
						return nil, err
					}
				}
			} else {
				// TODO: error details
				return nil, dsh_utils.NewError("target is not a map", map[string]any{
					"key": sourceKey,
				})
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
				if mergeKey, exist := merges[sourceKey]; exist {
					if mergeKey == projectConfigMergeTypeReplace {
						target[k] = sourceList
					} else if mergeKey == projectConfigMergeTypeInsert {
						target[k] = append(sourceList, targetList...)
					} else {
						target[k] = append(targetList, sourceList...)
					}
				} else {
					target[k] = append(targetList, sourceList...)
				}
			} else {
				// TODO: error details
				return nil, dsh_utils.NewError("target is not a list", map[string]any{
					"key": sourceKey,
				})
			}
			break
		default:
			target[k] = v
		}
	}
	return target, nil
}
