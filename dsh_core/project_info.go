package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"regexp"
)

type projectInfo struct {
	path         string
	name         string
	manifestPath string
	manifestType projectManifestType
	manifest     *projectManifest
}

type projectManifest struct {
	Name    string
	Runtime projectManifestRuntime
	Option  projectManifestOption
	Script  projectManifestScript
	Config  projectManifestConfig
}

type projectManifestRuntime struct {
	MinVersion dsh_utils.Version `yaml:"minVersion" toml:"minVersion" json:"minVersion"`
	MaxVersion dsh_utils.Version `yaml:"maxVersion" toml:"maxVersion" json:"maxVersion"`
}

type projectManifestOption struct {
	Items []projectManifestOptionItem
}

type projectManifestOptionItem struct {
	Name string
}

type projectManifestScript struct {
	Sources []projectManifestSource
	Imports []projectManifestImport
}

type projectManifestConfig struct {
	Sources []projectManifestSource
	Imports []projectManifestImport
}

type projectManifestSource struct {
	Dir   string
	Files []string
}

type projectManifestImport struct {
	Local *projectManifestImportLocal
	Git   *projectManifestImportGit
}

type projectManifestImportLocal struct {
	Dir string
}

type projectManifestImportGit struct {
	Url string
	Ref string
}

type projectManifestType string

const (
	projectManifestTypeYaml projectManifestType = "yaml"
	projectManifestTypeToml projectManifestType = "toml"
	projectManifestTypeJson projectManifestType = "json"
)

func loadProjectInfo(workspace *Workspace, path string) (project *projectInfo, err error) {
	manifestPath, manifestFileType := dsh_utils.SelectFile(path, []string{
		"project.yml",
		"project.yaml",
		"project.toml",
		"project.json",
	}, []dsh_utils.FileType{
		dsh_utils.FileTypeYaml,
		dsh_utils.FileTypeToml,
		dsh_utils.FileTypeJson,
	})
	if manifestPath == "" {
		return nil, dsh_utils.NewError("project manifest file not found", map[string]any{
			"path": path,
		})
	}
	var manifestType projectManifestType
	manifest := &projectManifest{}
	switch manifestFileType {
	case dsh_utils.FileTypeYaml:
		manifestType = projectManifestTypeYaml
		if err = dsh_utils.ReadYamlFile(manifestPath, manifest); err != nil {
			return nil, err
		}
	case dsh_utils.FileTypeToml:
		manifestType = projectManifestTypeToml
		if err = dsh_utils.ReadTomlFile(manifestPath, manifest); err != nil {
			return nil, err
		}
	case dsh_utils.FileTypeJson:
		manifestType = projectManifestTypeJson
		if err = dsh_utils.ReadJsonFile(manifestPath, manifest); err != nil {
			return nil, err
		}
	default:
		panic(fmt.Sprintf("project manifest file type not supported: path=%s, type=%s", manifestPath, manifestFileType))
	}
	project = &projectInfo{
		path:         path,
		manifestPath: manifestPath,
		manifestType: manifestType,
		manifest:     manifest,
	}
	if err = project.check(); err != nil {
		return nil, err
	}
	project.name = project.manifest.Name
	return project, nil
}

func (info *projectInfo) check() (err error) {
	manifest := info.manifest

	err = dsh_utils.CheckRuntimeVersion(manifest.Runtime.MinVersion, manifest.Runtime.MaxVersion)
	if err != nil {
		return err
	}

	if manifest.Name == "" {
		return dsh_utils.NewError("project manifest invalid", map[string]any{
			"manifestPath": info.manifestPath,
			"field":        "name",
			"reason":       "name is empty",
		})
	}
	if matched, _ := regexp.MatchString("^[a-z][a-z0-9_]*$", manifest.Name); !matched {
		return dsh_utils.NewError("project manifest invalid", map[string]any{
			"manifestPath": info.manifestPath,
			"field":        "name",
			"reason":       "name is invalid: " + manifest.Name,
		})
	}

	for i := 0; i < len(manifest.Script.Sources); i++ {
		src := manifest.Script.Sources[i]
		if src.Dir == "" {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.manifestPath,
				"field":        "script.sources",
				"reason":       "dir is empty",
			})
		}
	}
	for i := 0; i < len(manifest.Script.Imports); i++ {
		imp := manifest.Script.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.manifestPath,
				"field":        "script.imports",
				"reason":       "local and git are both nil",
			})
		} else if imp.Local != nil && imp.Git != nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.manifestPath,
				"field":        "script.imports",
				"reason":       "local and git are both not nil",
			})
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"manifestPath": info.manifestPath,
					"field":        "script.imports.local",
					"reason":       "dir is empty",
				})
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" || imp.Git.Ref == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"manifestPath": info.manifestPath,
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
				"manifestPath": info.manifestPath,
				"field":        "config.sources",
				"reason":       "dir is empty",
			})
		}
	}
	for i := 0; i < len(manifest.Config.Imports); i++ {
		imp := manifest.Config.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.manifestPath,
				"field":        "config.imports",
				"reason":       "local and git are both nil",
			})
		} else if imp.Local != nil && imp.Git != nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"manifestPath": info.manifestPath,
				"field":        "config.imports",
				"reason":       "local and git are both not nil",
			})
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"manifestPath": info.manifestPath,
					"field":        "config.imports.local",
					"reason":       "dir is empty",
				})
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" || imp.Git.Ref == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"manifestPath": info.manifestPath,
					"field":        "config.imports.git",
					"reason":       "url or ref is empty",
				})
			}
		}
	}
	return nil
}
