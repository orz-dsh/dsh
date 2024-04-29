package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
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
	Order  int64
	Match  string
	Config map[string]any
	match  *vm.Program
}

type projectInstanceConfigSourceType string

const (
	projectInstanceConfigSourceTypeYaml projectInstanceConfigSourceType = "yaml"
	projectInstanceConfigSourceTypeToml projectInstanceConfigSourceType = "toml"
	projectInstanceConfigSourceTypeJson projectInstanceConfigSourceType = "json"
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
			if content.Match != "" {
				if content.match, err = dsh_utils.CompileExpr(content.Match); err != nil {
					return dsh_utils.WrapError(err, "compile match expr failed", map[string]any{
						"sourcePath": source.sourcePath,
					})
				}
			}
			source.content = content
		}
	}
	return nil
}

func (content *projectInstanceConfigSourceContent) merge(target map[string]any) {
	mergeMap(target, content.Config)
}

func mergeMap(target map[string]any, source map[string]any) map[string]any {
	if target == nil {
		target = make(map[string]any)
	}
	for k, v := range source {
		if m, ok := v.(map[string]any); ok {
			tm, tok := target[k].(map[string]any)
			if !tok {
				if tm != nil {
					panic(fmt.Sprintf("target[%s] is not a map", k))
				}
				tm = make(map[string]any)
			}
			target[k] = mergeMap(tm, m)
		} else if a, ok := v.([]any); ok {
			ta, tok := target[k].([]any)
			if !tok {
				if ta != nil {
					panic(fmt.Sprintf("target[%s] is not an array", k))
				}
				ta = make([]any, 0)
			}
			target[k] = mergeArray(ta, a)
		} else {
			target[k] = v
		}
	}
	return target
}

func mergeArray(target []any, source []any) []any {
	for i := 0; i < len(source); i++ {
		v := source[i]
		if m, ok := v.(map[string]any); ok {
			target = append(target, mergeMap(make(map[string]any), m))
		} else if a, ok := v.([]any); ok {
			target = append(target, mergeArray(make([]any, 0), a))
		} else {
			target = append(target, v)
		}
	}
	return target
}
