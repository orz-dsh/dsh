package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
)

type Config struct {
	SourceContainer *ConfigSourceContainer
	ImportContainer *ShallowImportContainer
}

type ConfigSource struct {
	SourcePath string
	SourceName string
	Content    *ConfigSourceContent
}

type ConfigSourceContainer struct {
	SourceNameMap map[string]*ConfigSource
	YamlSources   []*ConfigSource
}

type ConfigSourceContent struct {
	Order  int64
	Config map[string]interface{}
}

func NewConfigSourceContainer() *ConfigSourceContainer {
	return &ConfigSourceContainer{
		SourceNameMap: make(map[string]*ConfigSource),
	}
}

func (container *ConfigSourceContainer) ScanSources(sourceDir string, includeFiles []string) error {
	yamlSourcePaths, err := dsh_utils.ScanConfigSources(sourceDir, includeFiles)
	if err != nil {
		return err
	}
	for i := 0; i < len(yamlSourcePaths); i++ {
		source := &ConfigSource{
			SourcePath: filepath.Join(sourceDir, yamlSourcePaths[i]),
			SourceName: yamlSourcePaths[i],
		}
		if existSource, exist := container.SourceNameMap[source.SourcePath]; exist {
			if existSource.SourcePath == source.SourcePath {
				continue
			}
			return dsh_utils.NewError("config source name is duplicated", map[string]interface{}{
				"sourceName":  source.SourceName,
				"sourcePath1": source.SourcePath,
				"sourcePath2": existSource.SourcePath,
			})
		}
		container.SourceNameMap[source.SourceName] = source
		container.YamlSources = append(container.YamlSources, source)
	}
	return nil
}

func (container *ConfigSourceContainer) LoadSources() (err error) {
	for i := 0; i < len(container.YamlSources); i++ {
		source := container.YamlSources[i]
		if source.Content == nil {
			content := &ConfigSourceContent{}
			if err = dsh_utils.ReadYaml(source.SourcePath, content); err != nil {
				return err
			}
			source.Content = content
		}
	}
	return nil
}

func (content *ConfigSourceContent) Merge(target map[string]interface{}) {
	MergeMap(target, content.Config)
}

func MergeMap(target map[string]interface{}, source map[string]interface{}) map[string]interface{} {
	if target == nil {
		target = make(map[string]interface{})
	}
	for k, v := range source {
		if m, ok := v.(map[string]interface{}); ok {
			tm, tok := target[k].(map[string]interface{})
			if !tok {
				if tm != nil {
					panic(fmt.Sprintf("target[%s] is not a map", k))
				}
				tm = make(map[string]interface{})
			}
			target[k] = MergeMap(tm, m)
		} else if a, ok := v.([]interface{}); ok {
			ta, tok := target[k].([]interface{})
			if !tok {
				if ta != nil {
					panic(fmt.Sprintf("target[%s] is not an array", k))
				}
				ta = make([]interface{}, 0)
			}
			target[k] = MergeArray(ta, a)
		} else {
			target[k] = v
		}
	}
	return target
}

func MergeArray(target []interface{}, source []interface{}) []interface{} {
	for i := 0; i < len(source); i++ {
		v := source[i]
		if m, ok := v.(map[string]interface{}); ok {
			target = append(target, MergeMap(make(map[string]interface{}), m))
		} else if a, ok := v.([]interface{}); ok {
			target = append(target, MergeArray(make([]interface{}, 0), a))
		} else {
			target = append(target, v)
		}
	}
	return target
}
