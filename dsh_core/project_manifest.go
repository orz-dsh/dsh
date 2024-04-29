package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
	"regexp"
	"slices"
)

var projectNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*$")

type projectManifest struct {
	Name         string
	Runtime      *projectManifestRuntime
	Option       *projectManifestOption
	Script       *projectManifestScript
	Config       *projectManifestConfig
	manifestPath string
	manifestType projectManifestType
	projectPath  string
}

type projectManifestRuntime struct {
	MinVersion dsh_utils.Version `yaml:"minVersion" toml:"minVersion" json:"minVersion"`
	MaxVersion dsh_utils.Version `yaml:"maxVersion" toml:"maxVersion" json:"maxVersion"`
}

type projectManifestOption struct {
	Items    []*projectManifestOptionItem
	Verifies []string
	verifies []*vm.Program
}

type projectManifestOptionItem struct {
	Name         string
	Type         projectManifestOptionItemType
	Choices      []string
	Default      *string
	Optional     bool
	Links        []*projectManifestOptionItemLink
	defaultValue any
}

type projectManifestOptionItemLink struct {
	Project string
	Option  string
	Mapping string
	mapping *vm.Program
}

type projectManifestScript struct {
	Sources []*projectManifestSource
	Imports []*projectManifestImport
}

type projectManifestConfig struct {
	Sources []*projectManifestSource
	Imports []*projectManifestImport
}

type projectManifestSource struct {
	Dir   string
	Files []string
	Match string
	match *vm.Program
}

type projectManifestImport struct {
	Local *projectManifestImportLocal
	Git   *projectManifestImportGit
	Match string
	match *vm.Program
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

type projectManifestOptionItemType string

const (
	projectManifestOptionItemTypeString  projectManifestOptionItemType = "string"
	projectManifestOptionItemTypeBool    projectManifestOptionItemType = "bool"
	projectManifestOptionItemTypeInteger projectManifestOptionItemType = "integer"
	projectManifestOptionItemTypeDecimal projectManifestOptionItemType = "decimal"
)

func loadProjectManifest(projectPath string) (manifest *projectManifest, err error) {
	manifestPath, manifestFileType := dsh_utils.SelectFile(projectPath, []string{
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
			"projectPath": projectPath,
		})
	}
	var manifestType projectManifestType
	manifest = &projectManifest{
		Runtime: &projectManifestRuntime{},
		Option:  &projectManifestOption{},
		Script:  &projectManifestScript{},
		Config:  &projectManifestConfig{},
	}
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
	manifest.manifestPath = manifestPath
	manifest.manifestType = manifestType
	manifest.projectPath = projectPath
	if err = manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (manifest *projectManifest) init() (err error) {
	if manifest.Name == "" {
		return dsh_utils.NewError("project manifest invalid", map[string]any{
			"path":   manifest.manifestPath,
			"field":  "name",
			"reason": "name is empty",
		})
	}
	if checked := projectNameCheckRegex.MatchString(manifest.Name); !checked {
		return dsh_utils.NewError("project manifest invalid", map[string]any{
			"path":   manifest.manifestPath,
			"field":  "name",
			"reason": "name is invalid: " + manifest.Name,
		})
	}

	err = dsh_utils.CheckRuntimeVersion(manifest.Runtime.MinVersion, manifest.Runtime.MaxVersion)
	if err != nil {
		return dsh_utils.WrapError(err, "project manifest invalid", map[string]any{
			"path":  manifest.manifestPath,
			"field": "runtime",
		})
	}

	for i := 0; i < len(manifest.Option.Items); i++ {
		option := manifest.Option.Items[i]
		if option.Name == "" {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("option.items[%d].name", i),
				"reason": "name is empty",
			})
		}
		if checked := projectNameCheckRegex.MatchString(option.Name); !checked {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("option.items[%d].name", i),
				"reason": "name is invalid: " + option.Name,
			})
		}
		if option.Type == "" {
			option.Type = projectManifestOptionItemTypeString
		}
		if option.Default != nil {
			defaultValue := *option.Default
			if len(option.Choices) > 0 && !slices.Contains(option.Choices, defaultValue) {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("option.items[%d].default", i),
					"reason": fmt.Sprintf("default not in choices: default=%s, choices=%s", defaultValue, option.Choices),
				})
			}

			switch option.Type {
			case projectManifestOptionItemTypeString:
				option.defaultValue = defaultValue
			case projectManifestOptionItemTypeBool:
				option.defaultValue = defaultValue == "true"
			case projectManifestOptionItemTypeInteger:
				value, err := dsh_utils.ParseInteger(defaultValue)
				if err != nil {
					return dsh_utils.WrapError(err, "project manifest invalid", map[string]any{
						"path":   manifest.manifestPath,
						"field":  fmt.Sprintf("option.items[%d].default", i),
						"reason": "default is invalid: " + defaultValue,
					})
				}
				option.defaultValue = value
			case projectManifestOptionItemTypeDecimal:
				value, err := dsh_utils.ParseDecimal(defaultValue)
				if err != nil {
					return dsh_utils.WrapError(err, "project manifest invalid", map[string]any{
						"path":   manifest.manifestPath,
						"field":  fmt.Sprintf("option.items[%d].default", i),
						"reason": "default is invalid: " + defaultValue,
					})
				}
				option.defaultValue = value
			default:
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("option.items[%d].type", i),
					"reason": "type is invalid: " + option.Type,
				})
			}
		}
		for j := 0; j < len(option.Links); j++ {
			if option.Links[j].Project == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("option.items[%d].links[%d].project", i, j),
					"reason": "project is empty",
				})
			}
			if option.Links[j].Project == manifest.Name {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("option.items[%d].links[%d].project", i, j),
					"reason": "can not link same project option",
				})
			}
			if option.Links[j].Option == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("option.items[%d].links[%d].option", i, j),
					"reason": "option is empty",
				})
			}
			if option.Links[j].Mapping != "" {
				option.Links[j].mapping, err = dsh_utils.CompileExpr(option.Links[j].Mapping)
				if err != nil {
					return dsh_utils.WrapError(err, "project manifest invalid", map[string]any{
						"path":   manifest.manifestPath,
						"field":  fmt.Sprintf("option.items[%d].links[%d].mapping", i, j),
						"reason": "mapping is invalid",
					})
				}
			}
		}
	}
	for i := 0; i < len(manifest.Option.Verifies); i++ {
		if manifest.Option.Verifies[i] == "" {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("option.verifies[%d]", i),
				"reason": "verify is empty",
			})
		}
		verify, err := dsh_utils.CompileExpr(manifest.Option.Verifies[i])
		if err != nil {
			return dsh_utils.WrapError(err, "project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("option.verifies[%d]", i),
				"reason": "verify is invalid",
			})
		}
		manifest.Option.verifies = append(manifest.Option.verifies, verify)
	}

	for i := 0; i < len(manifest.Script.Sources); i++ {
		src := manifest.Script.Sources[i]
		if src.Dir == "" {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("script.sources[%d].dir", i),
				"reason": "dir is empty",
			})
		}
		if src.Match != "" {
			src.match, err = dsh_utils.CompileExpr(src.Match)
			if err != nil {
				return dsh_utils.WrapError(err, "project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("script.sources[%d].match", i),
					"reason": "match is invalid",
				})
			}
		}
	}
	for i := 0; i < len(manifest.Script.Imports); i++ {
		imp := manifest.Script.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("script.imports[%d]", i),
				"reason": "local and git are both nil",
			})
		} else if imp.Local != nil && imp.Git != nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("script.imports[%d]", i),
				"reason": "local and git are both not nil",
			})
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("script.imports[%d].local.dir", i),
					"reason": "dir is empty",
				})
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" || imp.Git.Ref == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("script.imports[%d].git", i),
					"reason": "url or ref is empty",
				})
			}
		}
		if imp.Match != "" {
			imp.match, err = dsh_utils.CompileExpr(imp.Match)
			if err != nil {
				return dsh_utils.WrapError(err, "project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("script.imports[%d].match", i),
					"reason": "match is invalid",
				})
			}
		}
	}

	for i := 0; i < len(manifest.Config.Sources); i++ {
		src := manifest.Config.Sources[i]
		if src.Dir == "" {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("config.sources[%d].dir", i),
				"reason": "dir is empty",
			})
		}
		if src.Match != "" {
			src.match, err = dsh_utils.CompileExpr(src.Match)
			if err != nil {
				return dsh_utils.WrapError(err, "project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("config.sources[%d].match", i),
					"reason": "match is invalid",
				})
			}
		}
	}
	for i := 0; i < len(manifest.Config.Imports); i++ {
		imp := manifest.Config.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("config.imports[%d]", i),
				"reason": "local and git are both nil",
			})
		} else if imp.Local != nil && imp.Git != nil {
			return dsh_utils.NewError("project manifest invalid", map[string]any{
				"path":   manifest.manifestPath,
				"field":  fmt.Sprintf("config.imports[%d]", i),
				"reason": "local and git are both not nil",
			})
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("config.imports[%d].local.dir", i),
					"reason": "dir is empty",
				})
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" || imp.Git.Ref == "" {
				return dsh_utils.NewError("project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("config.imports[%d].git", i),
					"reason": "url or ref is empty",
				})
			}
		}
		if imp.Match != "" {
			imp.match, err = dsh_utils.CompileExpr(imp.Match)
			if err != nil {
				return dsh_utils.WrapError(err, "project manifest invalid", map[string]any{
					"path":   manifest.manifestPath,
					"field":  fmt.Sprintf("config.imports[%d].match", i),
					"reason": "match is invalid",
				})
			}
		}
	}
	return nil
}
