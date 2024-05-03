package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
	"net/url"
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
	Name               string
	Type               projectManifestOptionItemType
	Choices            []string
	Default            *string
	Optional           bool
	Assigns            []*projectManifestOptionItemAssign
	defaultRawValue    string
	defaultParsedValue any
}

type projectManifestOptionItemAssign struct {
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
	url *url.URL
	ref *gitRef
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

func loadProjectManifest(projectPath string) (pm *projectManifest, err error) {
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
		return nil, errN("load project manifest error",
			reason("manifest file not found"),
			kv("projectPath", projectPath),
		)
	}
	pm = &projectManifest{
		Runtime: &projectManifestRuntime{},
		Option:  &projectManifestOption{},
		Script:  &projectManifestScript{},
		Config:  &projectManifestConfig{},
	}
	var manifestType projectManifestType
	switch manifestFileType {
	case dsh_utils.FileTypeYaml:
		manifestType = projectManifestTypeYaml
		err = dsh_utils.ReadYamlFile(manifestPath, pm)
	case dsh_utils.FileTypeToml:
		manifestType = projectManifestTypeToml
		err = dsh_utils.ReadTomlFile(manifestPath, pm)
	case dsh_utils.FileTypeJson:
		manifestType = projectManifestTypeJson
		err = dsh_utils.ReadJsonFile(manifestPath, pm)
	default:
		// impossible
		panic(desc("project manifest file type not supported",
			kv("manifestPath", manifestPath),
			kv("manifestFileType", manifestFileType),
		))
	}
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("read manifest file error"),
			kv("manifestPath", manifestPath),
		)
	}
	pm.manifestPath = manifestPath
	pm.manifestType = manifestType
	pm.projectPath = projectPath
	if err = pm.init(); err != nil {
		return nil, err
	}
	return pm, nil
}

func (pm *projectManifest) init() (err error) {
	if pm.Name == "" {
		return errN("project manifest invalid",
			reason("name empty"),
			kv("path", pm.manifestPath),
			kv("field", "name"),
		)
	}
	if checked := projectNameCheckRegex.MatchString(pm.Name); !checked {
		return errN("project manifest invalid",
			reason("value invalid"),
			kv("path", pm.manifestPath),
			kv("field", "name"),
			kv("value", pm.Name),
		)
	}

	err = dsh_utils.CheckRuntimeVersion(pm.Runtime.MinVersion, pm.Runtime.MaxVersion)
	if err != nil {
		return errW(err, "project manifest invalid",
			reason("runtime incompatible"),
			kv("path", pm.manifestPath),
			kv("field", "runtime"),
		)
	}

	optionsByName := make(map[string]bool)
	for i := 0; i < len(pm.Option.Items); i++ {
		option := pm.Option.Items[i]
		if option.Name == "" {
			return errN("project manifest invalid",
				reason("name empty"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].name", i)),
			)
		}
		if checked := projectNameCheckRegex.MatchString(option.Name); !checked {
			return errN("project manifest invalid",
				reason("value invalid"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].name", i)),
				kv("value", option.Name),
			)
		}
		if _, exist := optionsByName[option.Name]; exist {
			return errN("project manifest invalid",
				reason("name duplicated"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].name", i)),
				kv("value", option.Name),
			)
		}
		optionsByName[option.Name] = true
		if option.Type == "" {
			option.Type = projectManifestOptionItemTypeString
		}
		if option.Default != nil {
			switch option.Type {
			case projectManifestOptionItemTypeString:
			case projectManifestOptionItemTypeBool:
			case projectManifestOptionItemTypeInteger:
			case projectManifestOptionItemTypeDecimal:
			default:
				return errN("project manifest invalid",
					reason("value invalid"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].type", i)),
					kv("value", option.Type),
				)
			}

			defaultRawValue := *option.Default
			defaultParsedValue, err := option.parseValue(defaultRawValue)
			if err != nil {
				return errW(err, "project manifest invalid",
					reason("value invalid"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].default", i)),
					kv("value", defaultRawValue),
				)
			}
			option.defaultRawValue = defaultRawValue
			option.defaultParsedValue = defaultParsedValue
		}
		assignsByTarget := make(map[string]bool)
		for j := 0; j < len(option.Assigns); j++ {
			if option.Assigns[j].Project == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", i, j)),
				)
			}
			if option.Assigns[j].Project == pm.Name {
				return errN("project manifest invalid",
					reason("can not assign to self project option"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", i, j)),
				)
			}
			if option.Assigns[j].Option == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d].option", i, j)),
				)
			}
			assignTarget := option.Assigns[j].Project + "." + option.Assigns[j].Option
			if _, exists := assignsByTarget[assignTarget]; exists {
				return errN("project manifest invalid",
					reason("option assign target duplicated"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d]", i, j)),
					kv("target", assignTarget),
				)
			}
			assignsByTarget[assignTarget] = true
			if option.Assigns[j].Mapping != "" {
				option.Assigns[j].mapping, err = dsh_utils.CompileExpr(option.Assigns[j].Mapping)
				if err != nil {
					return errW(err, "project manifest invalid",
						reason("value invalid"),
						kv("path", pm.manifestPath),
						kv("field", fmt.Sprintf("option.items[%d].assigns[%d].mapping", i, j)),
						kv("value", option.Assigns[j].Mapping),
					)
				}
			}
		}
	}
	for i := 0; i < len(pm.Option.Verifies); i++ {
		if pm.Option.Verifies[i] == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
			)
		}
		verify, err := dsh_utils.CompileExpr(pm.Option.Verifies[i])
		if err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
				kv("value", pm.Option.Verifies[i]),
			)
		}
		pm.Option.verifies = append(pm.Option.verifies, verify)
	}

	for i := 0; i < len(pm.Script.Sources); i++ {
		src := pm.Script.Sources[i]
		if src.Dir == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("script.sources[%d].dir", i)),
			)
		}
		if src.Match != "" {
			src.match, err = dsh_utils.CompileExpr(src.Match)
			if err != nil {
				return errW(err, "project manifest invalid",
					reason("value invalid"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("script.sources[%d].match", i)),
					kv("value", src.Match),
				)
			}
		}
	}
	for i := 0; i < len(pm.Script.Imports); i++ {
		imp := pm.Script.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return errN("project manifest invalid",
				reason("local and git are both nil"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("script.imports[%d]", i)),
			)
		} else if imp.Local != nil && imp.Git != nil {
			return errN("project manifest invalid",
				reason("local and git are both not nil"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("script.imports[%d]", i)),
			)
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("script.imports[%d].local.dir", i)),
				)
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("script.imports[%d].git.url", i)),
				)
			}
			if imp.Git.Ref == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("script.imports[%d].git.ref", i)),
				)
			}
			if imp.Git.url, err = url.Parse(imp.Git.Url); err != nil {
				return errW(err, "project manifest invalid",
					reason("value invalid"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("script.imports[%d].git.url", i)),
					kv("value", imp.Git.Url),
				)
			}
			imp.Git.ref = parseGitRef(imp.Git.Ref)
		}
		if imp.Match != "" {
			imp.match, err = dsh_utils.CompileExpr(imp.Match)
			if err != nil {
				return errW(err, "project manifest invalid",
					reason("value invalid"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("script.imports[%d].match", i)),
					kv("value", imp.Match),
				)
			}
		}
	}

	for i := 0; i < len(pm.Config.Sources); i++ {
		src := pm.Config.Sources[i]
		if src.Dir == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("config.sources[%d].dir", i)),
			)
		}
		if src.Match != "" {
			src.match, err = dsh_utils.CompileExpr(src.Match)
			if err != nil {
				return errW(err, "project manifest invalid",
					reason("value invalid"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("config.sources[%d].match", i)),
					kv("value", src.Match),
				)
			}
		}
	}
	for i := 0; i < len(pm.Config.Imports); i++ {
		imp := pm.Config.Imports[i]
		if imp.Local == nil && imp.Git == nil {
			return errN("project manifest invalid",
				reason("local and git are both nil"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("config.imports[%d]", i)),
			)
		} else if imp.Local != nil && imp.Git != nil {
			return errN("project manifest invalid",
				reason("local and git are both not nil"),
				kv("path", pm.manifestPath),
				kv("field", fmt.Sprintf("config.imports[%d]", i)),
			)
		} else if imp.Local != nil {
			if imp.Local.Dir == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("config.imports[%d].local.dir", i)),
				)
			}
		} else if imp.Git != nil {
			if imp.Git.Url == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("config.imports[%d].git.url", i)),
				)
			}
			if imp.Git.Ref == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("config.imports[%d].git.ref", i)),
				)
			}
			if imp.Git.url, err = url.Parse(imp.Git.Url); err != nil {
				return errW(err, "project manifest invalid",
					reason("value invalid"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("script.imports[%d].git.url", i)),
					kv("value", imp.Git.Url),
				)
			}
			imp.Git.ref = parseGitRef(imp.Git.Ref)
		}
		if imp.Match != "" {
			imp.match, err = dsh_utils.CompileExpr(imp.Match)
			if err != nil {
				return errW(err, "project manifest invalid",
					reason("value invalid"),
					kv("path", pm.manifestPath),
					kv("field", fmt.Sprintf("config.imports[%d].match", i)),
					kv("value", imp.Match),
				)
			}
		}
	}
	return nil
}

func (item *projectManifestOptionItem) parseValue(rawValue string) (any, error) {
	if len(item.Choices) > 0 && !slices.Contains(item.Choices, rawValue) {
		return nil, errN("option parse value error",
			reason("not in choices"),
			kv("name", item.Name),
			kv("value", rawValue),
			kv("choices", item.Choices),
		)
	}
	var parsedValue any = nil
	switch item.Type {
	case projectManifestOptionItemTypeString:
		parsedValue = rawValue
	case projectManifestOptionItemTypeBool:
		parsedValue = rawValue == "true"
	case projectManifestOptionItemTypeInteger:
		integer, err := dsh_utils.ParseInteger(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse integer error"),
				kv("name", item.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = integer
	case projectManifestOptionItemTypeDecimal:
		decimal, err := dsh_utils.ParseDecimal(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse decimal error"),
				kv("name", item.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = decimal
	default:
		// impossible
		panic(desc("option type not supported",
			kv("optionName", item.Name),
			kv("optionType", item.Type),
		))
	}
	return parsedValue, nil
}
