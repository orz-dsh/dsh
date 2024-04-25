package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
)

type Manifest struct {
	Name    string
	Runtime ManifestRuntime
	Script  ManifestScript
	Config  ManifestConfig
}

type ManifestRuntime struct {
	MinVersion dsh_utils.Version `yaml:"minVersion"`
	MaxVersion dsh_utils.Version `yaml:"maxVersion"`
}

type ManifestScript struct {
	Sources []ManifestSource
	Imports []ManifestImport
}

type ManifestConfig struct {
	Sources []ManifestSource
	Imports []ManifestImport
}

type ManifestSource struct {
	Dir   string
	Files []string
}

type ManifestImport struct {
	Local *ManifestImportLocal
	Git   *ManifestImportGit
}

type ManifestImportLocal struct {
	Dir string
}

type ManifestImportGit struct {
	Url string
	Ref string
}

func (manifest *Manifest) LoadYaml(yamlPath string) (err error) {
	err = dsh_utils.ReadYaml(yamlPath, manifest)
	if err != nil {
		return err
	}
	if manifest.Name == "" {
		return dsh_utils.NewError("manifest name is empty", map[string]interface{}{
			"yamlPath": yamlPath,
		})
	}
	return nil
}

func (manifest *Manifest) PreCheck(project *Project) (err error) {
	err = dsh_utils.CheckRuntimeVersion(manifest.Runtime.MinVersion, manifest.Runtime.MaxVersion)
	if err != nil {
		return err
	}
	project.Name = manifest.Name
	return nil
}

func (manifest *Manifest) Setup(project *Project) (err error) {
	for i := 0; i < len(manifest.Script.Sources); i++ {
		src := manifest.Script.Sources[i]
		if src.Dir == "" {
			return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
				"projectPath": project.Path,
				"field":       "script.sources.dir",
				"reason":      "dir is empty",
			})
		}
		if err = project.ScanScriptSources(filepath.Join(project.Path, src.Dir), src.Files); err != nil {
			return err
		}
	}
	for i := 0; i < len(manifest.Script.Imports); i++ {
		imp := manifest.Script.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
				"projectPath": project.Path,
				"field":       "script.imports",
				"reason":      "local and git are both nil",
			})
		} else if imp.Local != nil && imp.Git != nil {
			return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
				"projectPath": project.Path,
				"field":       "script.imports",
				"reason":      "local and git are both not nil",
			})
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
					"projectPath": project.Path,
					"field":       "script.imports.local.dir",
					"reason":      "dir is empty",
				})
			}
			if err = project.AddLocalImport(ImportScopeScript, imp.Local.Dir); err != nil {
				return err
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" || imp.Git.Ref == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
					"projectPath": project.Path,
					"field":       "script.imports.git",
					"reason":      "url or ref is empty",
				})
			}
			if err = project.AddGitImport(ImportScopeScript, imp.Git.Url, imp.Git.Ref); err != nil {
				return err
			}
		}
	}
	for i := 0; i < len(manifest.Config.Sources); i++ {
		src := manifest.Config.Sources[i]
		if src.Dir == "" {
			return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
				"projectPath": project.Path,
				"field":       "config.sources.dir",
				"reason":      "dir is empty",
			})
		}
		if err = project.ScanConfigSources(filepath.Join(project.Path, src.Dir), src.Files); err != nil {
			return err
		}
	}
	for i := 0; i < len(manifest.Config.Imports); i++ {
		imp := manifest.Config.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
				"projectPath": project.Path,
				"field":       "config.imports",
				"reason":      "local and git are both nil",
			})
		} else if imp.Local != nil && imp.Git != nil {
			return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
				"projectPath": project.Path,
				"field":       "config.imports",
				"reason":      "local and git are both not nil",
			})
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
					"projectPath": project.Path,
					"field":       "config.imports.local.dir",
					"reason":      "dir is empty",
				})
			}
			if err = project.AddLocalImport(ImportScopeConfig, imp.Local.Dir); err != nil {
				return err
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" || imp.Git.Ref == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]interface{}{
					"projectPath": project.Path,
					"field":       "config.imports.git",
					"reason":      "url or ref is empty",
				})
			}
			if err = project.AddGitImport(ImportScopeConfig, imp.Git.Url, imp.Git.Ref); err != nil {
				return err
			}
		}
	}
	return nil
}
