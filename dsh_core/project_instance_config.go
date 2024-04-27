package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
)

type ProjectInstanceConfig struct {
	SourceContainer *ProjectInstanceConfigSourceContainer
	ImportContainer *ProjectInstanceImportShallowContainer
}

type ProjectInstanceConfigSource struct {
	SourcePath string
	SourceName string
	SourceType ProjectInstanceConfigSourceType
	Content    *ProjectInstanceConfigSourceContent
}

type ProjectInstanceConfigSourceContainer struct {
	Context       *Context
	SourceNameMap map[string]*ProjectInstanceConfigSource
	Sources       []*ProjectInstanceConfigSource
}

type ProjectInstanceConfigSourceContent struct {
	Order  int64
	Config map[string]any
}

type ProjectInstanceConfigSourceType string

const (
	ProjectInstanceConfigSourceTypeYaml ProjectInstanceConfigSourceType = "yaml"
	ProjectInstanceConfigSourceTypeToml ProjectInstanceConfigSourceType = "toml"
	ProjectInstanceConfigSourceTypeJson ProjectInstanceConfigSourceType = "json"
)

func NewProjectInstanceConfig(context *Context) *ProjectInstanceConfig {
	return &ProjectInstanceConfig{
		SourceContainer: &ProjectInstanceConfigSourceContainer{
			Context:       context,
			SourceNameMap: make(map[string]*ProjectInstanceConfigSource),
		},
		ImportContainer: NewShallowImportContainer(context, ProjectInstanceImportScopeConfig),
	}
}

func (container *ProjectInstanceConfigSourceContainer) ScanSources(sourceDir string, includeFiles []string) error {
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
		var sourceType ProjectInstanceConfigSourceType
		switch fileType {
		case dsh_utils.FileTypeYaml:
			sourceType = ProjectInstanceConfigSourceTypeYaml
		case dsh_utils.FileTypeToml:
			sourceType = ProjectInstanceConfigSourceTypeToml
		case dsh_utils.FileTypeJson:
			sourceType = ProjectInstanceConfigSourceTypeJson
		default:
			container.Context.Logger.Panic("unsupported config source type", map[string]any{
				"filePath": filePath,
				"fileType": fileType,
			})
			continue
		}
		source := &ProjectInstanceConfigSource{
			SourcePath: filepath.Join(sourceDir, filePath),
			SourceName: dsh_utils.RemoveFileExt(filePath),
			SourceType: sourceType,
		}
		if existSource, exist := container.SourceNameMap[source.SourcePath]; exist {
			if existSource.SourcePath == source.SourcePath {
				continue
			}
			return dsh_utils.NewError("config source name is duplicated", map[string]any{
				"sourceName":  source.SourceName,
				"sourcePath1": source.SourcePath,
				"sourcePath2": existSource.SourcePath,
			})
		}
		container.SourceNameMap[source.SourceName] = source
		container.Sources = append(container.Sources, source)
	}
	return nil
}

func (container *ProjectInstanceConfigSourceContainer) LoadSources() (err error) {
	for i := 0; i < len(container.Sources); i++ {
		source := container.Sources[i]
		if source.Content == nil {
			content := &ProjectInstanceConfigSourceContent{}
			if source.SourceType == ProjectInstanceConfigSourceTypeYaml {
				if err = dsh_utils.ReadYamlFile(source.SourcePath, content); err != nil {
					return err
				}
			} else if source.SourceType == ProjectInstanceConfigSourceTypeToml {
				if err = dsh_utils.ReadTomlFile(source.SourcePath, content); err != nil {
					return err
				}
			} else if source.SourceType == ProjectInstanceConfigSourceTypeJson {
				if err = dsh_utils.ReadJsonFile(source.SourcePath, content); err != nil {
					return err
				}
			} else {
				container.Context.Logger.Panic("unsupported config source type", map[string]any{
					"sourcePath": source.SourcePath,
				})
				return nil
			}
			source.Content = content
		}
	}
	return nil
}

func (content *ProjectInstanceConfigSourceContent) Merge(target map[string]any) {
	MergeMap(target, content.Config)
}

func MergeMap(target map[string]any, source map[string]any) map[string]any {
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
			target[k] = MergeMap(tm, m)
		} else if a, ok := v.([]any); ok {
			ta, tok := target[k].([]any)
			if !tok {
				if ta != nil {
					panic(fmt.Sprintf("target[%s] is not an array", k))
				}
				ta = make([]any, 0)
			}
			target[k] = MergeArray(ta, a)
		} else {
			target[k] = v
		}
	}
	return target
}

func MergeArray(target []any, source []any) []any {
	for i := 0; i < len(source); i++ {
		v := source[i]
		if m, ok := v.(map[string]any); ok {
			target = append(target, MergeMap(make(map[string]any), m))
		} else if a, ok := v.([]any); ok {
			target = append(target, MergeArray(make([]any, 0), a))
		} else {
			target = append(target, v)
		}
	}
	return target
}
