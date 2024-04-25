package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"text/template"
)

type Script struct {
	SourceContainer *ScriptSourceContainer
	ImportContainer *ShallowImportContainer
}

type ScriptSource struct {
	SourcePath string
	SourceName string
}

type ScriptSourceContainer struct {
	SourceNameMap      map[string]*ScriptSource
	PlainSources       []*ScriptSource
	TemplateSources    []*ScriptSource
	TemplateLibSources []*ScriptSource
}

func NewScriptSourceContainer() *ScriptSourceContainer {
	return &ScriptSourceContainer{
		SourceNameMap: make(map[string]*ScriptSource),
	}
}

func (sc *ScriptSourceContainer) ScanSources(sourceDir string, includeFiles []string) error {
	plainSourcePaths, templateSourcePaths, templateLibSourcePaths, err := dsh_utils.ScanScriptSources(sourceDir, includeFiles)
	if err != nil {
		return err
	}
	for i := 0; i < len(plainSourcePaths); i++ {
		source := &ScriptSource{
			SourcePath: filepath.Join(sourceDir, plainSourcePaths[i]),
			SourceName: plainSourcePaths[i],
		}
		if existSource, exist := sc.SourceNameMap[source.SourceName]; exist {
			if existSource.SourcePath == source.SourcePath {
				continue
			}
			return dsh_utils.NewError("script source name is duplicated", map[string]interface{}{
				"sourceName":  source.SourceName,
				"sourcePath1": source.SourcePath,
				"sourcePath2": existSource.SourcePath,
			})
		}
		sc.SourceNameMap[source.SourceName] = source
		sc.PlainSources = append(sc.PlainSources, source)
	}
	for i := 0; i < len(templateSourcePaths); i++ {
		source := &ScriptSource{
			SourcePath: filepath.Join(sourceDir, templateSourcePaths[i]),
			SourceName: templateSourcePaths[i][:len(templateSourcePaths[i])-len(".dtpl")],
		}
		if existSource, exist := sc.SourceNameMap[source.SourceName]; exist {
			if existSource.SourcePath == source.SourcePath {
				continue
			}
			return dsh_utils.NewError("script source name is duplicated", map[string]interface{}{
				"sourceName":  source.SourceName,
				"sourcePath1": source.SourcePath,
				"sourcePath2": existSource.SourcePath,
			})
		}
		sc.SourceNameMap[source.SourceName] = source
		sc.TemplateSources = append(sc.TemplateSources, source)
	}
	for i := 0; i < len(templateLibSourcePaths); i++ {
		source := &ScriptSource{
			SourcePath: filepath.Join(sourceDir, templateLibSourcePaths[i]),
			SourceName: templateLibSourcePaths[i],
		}
		if existSource, exist := sc.SourceNameMap[source.SourceName]; exist {
			if existSource.SourcePath == source.SourcePath {
				continue
			}
			return dsh_utils.NewError("script source name is duplicated", map[string]interface{}{
				"sourceName":  source.SourceName,
				"sourcePath1": source.SourcePath,
				"sourcePath2": existSource.SourcePath,
			})
		}
		sc.SourceNameMap[source.SourceName] = source
		sc.TemplateLibSources = append(sc.TemplateLibSources, source)
	}
	return nil
}

func (sc *ScriptSourceContainer) BuildSources(config map[string]interface{}, funcs template.FuncMap, outputPath string) (err error) {
	for i := 0; i < len(sc.PlainSources); i++ {
		source := sc.PlainSources[i]
		outputTargetPath := filepath.Join(outputPath, source.SourceName)
		// TODO: logging
		fmt.Printf("copy %s to %s\n", source.SourcePath, outputTargetPath)
		err = dsh_utils.LinkOrCopyFile(source.SourcePath, outputTargetPath)
		if err != nil {
			return err
		}
	}

	var templateLibSourcePaths []string
	for i := 0; i < len(sc.TemplateLibSources); i++ {
		templateLibSourcePaths = append(templateLibSourcePaths, sc.TemplateLibSources[i].SourcePath)
	}
	for i := 0; i < len(sc.TemplateSources); i++ {
		source := sc.TemplateSources[i]
		outputTargetPath := filepath.Join(outputPath, source.SourceName)
		// TODO: logging
		fmt.Printf("generate %s to %s\n", source.SourcePath, outputTargetPath)
		if err = BuildTemplate(config, funcs, source.SourcePath, templateLibSourcePaths, outputTargetPath); err != nil {
			return err
		}
	}

	return nil
}
