package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
)

type ProjectInfo struct {
	Path         string
	Name         string
	ManifestPath string
	ManifestType ProjectManifestType
	Manifest     *ProjectManifest
}

type ProjectManifest struct {
	Name    string
	Runtime ProjectManifestRuntime
	Option  ProjectManifestOption
	Script  ProjectManifestScript
	Config  ProjectManifestConfig
}

type ProjectManifestRuntime struct {
	MinVersion dsh_utils.Version `yaml:"minVersion"`
	MaxVersion dsh_utils.Version `yaml:"maxVersion"`
}

type ProjectManifestOption struct {
	Items []ProjectManifestOptionItem
}

type ProjectManifestOptionItem struct {
	Name string
}

type ProjectManifestScript struct {
	Sources []ProjectManifestSource
	Imports []ProjectManifestImport
}

type ProjectManifestConfig struct {
	Sources []ProjectManifestSource
	Imports []ProjectManifestImport
}

type ProjectManifestSource struct {
	Dir   string
	Files []string
}

type ProjectManifestImport struct {
	Local *ProjectManifestImportLocal
	Git   *ProjectManifestImportGit
}

type ProjectManifestImportLocal struct {
	Dir string
}

type ProjectManifestImportGit struct {
	Url string
	Ref string
}

type ProjectManifestType string

const (
	ProjectManifestTypeYaml ProjectManifestType = "yaml"
	ProjectManifestTypeToml ProjectManifestType = "toml"
	ProjectManifestTypeJson ProjectManifestType = "json"
)

func LoadProjectInfo(workspace *Workspace, path string) (project *ProjectInfo, err error) {
	manifestPath := dsh_utils.SelectFiles(path, []string{"project.yml", "project.yaml", "project.toml", "project.json"})
	if manifestPath == "" {
		return nil, dsh_utils.NewError("project manifest file not found", map[string]any{
			"path": path,
		})
	}
	var manifestType ProjectManifestType
	manifest := &ProjectManifest{}
	if dsh_utils.IsYamlFile(manifestPath) {
		manifestType = ProjectManifestTypeYaml
		if err = dsh_utils.ReadYamlFile(manifestPath, manifest); err != nil {
			return nil, err
		}
	} else if dsh_utils.IsTomlFile(manifestPath) {
		manifestType = ProjectManifestTypeToml
		if err = dsh_utils.ReadTomlFile(manifestPath, manifest); err != nil {
			return nil, err
		}
	} else if dsh_utils.IsJsonFile(manifestPath) {
		manifestType = ProjectManifestTypeJson
		if err = dsh_utils.ReadJsonFile(manifestPath, manifest); err != nil {
			return nil, err
		}
	} else {
		workspace.Logger.Panic("project manifest file type not supported: path=%s", manifestPath)
		return nil, nil
	}
	project = &ProjectInfo{
		Path:         path,
		ManifestPath: manifestPath,
		ManifestType: manifestType,
		Manifest:     manifest,
	}
	if err = project.Check(); err != nil {
		return nil, err
	}
	project.Name = project.Manifest.Name
	return project, nil
}

func (info *ProjectInfo) Check() (err error) {
	manifest := info.Manifest

	if manifest.Name == "" {
		return dsh_utils.NewError("project manifest invalid", map[string]any{
			"manifestPath": info.ManifestPath,
			"field":        "name",
			"reason":       "name is empty",
		})
	}
	if matched, _ := regexp.MatchString("^[a-z][a-z0-9_]*$", manifest.Name); !matched {
		return dsh_utils.NewError("project manifest invalid", map[string]any{
			"manifestPath": info.ManifestPath,
			"field":        "name",
			"reason":       "name is invalid: " + manifest.Name,
		})
	}

	for i := 0; i < len(manifest.Script.Sources); i++ {
		src := manifest.Script.Sources[i]
		if src.Dir == "" {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.ManifestPath,
				"field":        "script.sources",
				"reason":       "dir is empty",
			})
		}
	}
	for i := 0; i < len(manifest.Script.Imports); i++ {
		imp := manifest.Script.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.ManifestPath,
				"field":        "script.imports",
				"reason":       "local and git are both nil",
			})
		} else if imp.Local != nil && imp.Git != nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.ManifestPath,
				"field":        "script.imports",
				"reason":       "local and git are both not nil",
			})
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"manifestPath": info.ManifestPath,
					"field":        "script.imports.local",
					"reason":       "dir is empty",
				})
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" || imp.Git.Ref == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"manifestPath": info.ManifestPath,
					"field":        "script.imports.git",
					"reason":       "url or ref is empty",
				})
			}
		}
	}
	for i := 0; i < len(manifest.Config.Sources); i++ {
		src := manifest.Config.Sources[i]
		if src.Dir == "" {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.ManifestPath,
				"field":        "config.sources",
				"reason":       "dir is empty",
			})
		}
	}
	for i := 0; i < len(manifest.Config.Imports); i++ {
		imp := manifest.Config.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.ManifestPath,
				"field":        "config.imports",
				"reason":       "local and git are both nil",
			})
		} else if imp.Local != nil && imp.Git != nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.ManifestPath,
				"field":        "config.imports",
				"reason":       "local and git are both not nil",
			})
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"manifestPath": info.ManifestPath,
					"field":        "config.imports.local",
					"reason":       "dir is empty",
				})
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" || imp.Git.Ref == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"manifestPath": info.ManifestPath,
					"field":        "config.imports.git",
					"reason":       "url or ref is empty",
				})
			}
		}
	}
	return nil
}
