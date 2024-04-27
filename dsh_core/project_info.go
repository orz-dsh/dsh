package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
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

func LoadProjectInfo(path string) (project *ProjectInfo, err error) {
	manifestYamlPath := filepath.Join(path, "project.yml")
	if !dsh_utils.IsFileExists(manifestYamlPath) {
		manifestYamlPath = filepath.Join(path, "project.yaml")
		if !dsh_utils.IsFileExists(manifestYamlPath) {
			manifestYamlPath = ""
		}
	}
	var manifestPath string
	var manifestType ProjectManifestType
	if manifestYamlPath != "" {
		manifestPath = manifestYamlPath
		manifestType = ProjectManifestTypeYaml
	} else {
		return nil, dsh_utils.NewError("project manifest file not found", map[string]any{
			"path": path,
		})
	}

	manifest := &ProjectManifest{}
	if manifestType == ProjectManifestTypeYaml {
		if err = dsh_utils.ReadYaml(manifestPath, manifest); err != nil {
			return nil, err
		}
	} else if manifestType == ProjectManifestTypeToml {
		// TODO
		panic("toml not supported yet")
	} else if manifestType == ProjectManifestTypeJson {
		// TODO
		panic("json not supported yet")
	} else {
		panic("unsupported manifest type: " + manifestType)
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
