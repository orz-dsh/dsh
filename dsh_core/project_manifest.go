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
	manifestType manifestMetadataType
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
	Registry *projectManifestImportRegistry
	Local    *projectManifestImportLocal
	Git      *projectManifestImportGit
	Match    string
	match    *vm.Program
}

type projectManifestImportRegistry struct {
	Name string
	Path string
	Ref  string
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

type projectManifestOptionItemType string

const (
	projectManifestOptionItemTypeString  projectManifestOptionItemType = "string"
	projectManifestOptionItemTypeBool    projectManifestOptionItemType = "bool"
	projectManifestOptionItemTypeInteger projectManifestOptionItemType = "integer"
	projectManifestOptionItemTypeDecimal projectManifestOptionItemType = "decimal"
)

func loadProjectManifest(projectPath string) (manifest *projectManifest, err error) {
	manifest = &projectManifest{
		Runtime: &projectManifestRuntime{},
		Option:  &projectManifestOption{},
		Script:  &projectManifestScript{},
		Config:  &projectManifestConfig{},
	}
	metadata, err := loadManifest(projectPath, []string{"project"}, manifest, true)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("load manifest error"),
			kv("projectPath", projectPath),
		)
	}
	manifest.manifestPath = metadata.manifestPath
	manifest.manifestType = metadata.manifestType
	manifest.projectPath = projectPath
	if err = manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *projectManifest) init() (err error) {
	if m.Name == "" {
		return errN("project manifest invalid",
			reason("name empty"),
			kv("path", m.manifestPath),
			kv("field", "name"),
		)
	}
	if checked := projectNameCheckRegex.MatchString(m.Name); !checked {
		return errN("project manifest invalid",
			reason("value invalid"),
			kv("path", m.manifestPath),
			kv("field", "name"),
			kv("value", m.Name),
		)
	}

	err = dsh_utils.CheckRuntimeVersion(m.Runtime.MinVersion, m.Runtime.MaxVersion)
	if err != nil {
		return errW(err, "project manifest invalid",
			reason("runtime incompatible"),
			kv("path", m.manifestPath),
			kv("field", "runtime"),
		)
	}

	optionNamesDict := make(map[string]bool)
	for i := 0; i < len(m.Option.Items); i++ {
		option := m.Option.Items[i]
		if option.Name == "" {
			return errN("project manifest invalid",
				reason("name empty"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].name", i)),
			)
		}
		if checked := projectNameCheckRegex.MatchString(option.Name); !checked {
			return errN("project manifest invalid",
				reason("value invalid"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].name", i)),
				kv("value", option.Name),
			)
		}
		if _, exist := optionNamesDict[option.Name]; exist {
			return errN("project manifest invalid",
				reason("name duplicated"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].name", i)),
				kv("value", option.Name),
			)
		}
		optionNamesDict[option.Name] = true
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
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].type", i)),
					kv("value", option.Type),
				)
			}

			defaultRawValue := *option.Default
			defaultParsedValue, err := option.parseValue(defaultRawValue)
			if err != nil {
				return errW(err, "project manifest invalid",
					reason("value invalid"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].default", i)),
					kv("value", defaultRawValue),
				)
			}
			option.defaultRawValue = defaultRawValue
			option.defaultParsedValue = defaultParsedValue
		}
		assignTargetsDict := make(map[string]bool)
		for j := 0; j < len(option.Assigns); j++ {
			if option.Assigns[j].Project == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", i, j)),
				)
			}
			if option.Assigns[j].Project == m.Name {
				return errN("project manifest invalid",
					reason("can not assign to self project option"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", i, j)),
				)
			}
			if option.Assigns[j].Option == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d].option", i, j)),
				)
			}
			assignTarget := option.Assigns[j].Project + "." + option.Assigns[j].Option
			if _, exists := assignTargetsDict[assignTarget]; exists {
				return errN("project manifest invalid",
					reason("option assign target duplicated"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d]", i, j)),
					kv("target", assignTarget),
				)
			}
			assignTargetsDict[assignTarget] = true
			if option.Assigns[j].Mapping != "" {
				option.Assigns[j].mapping, err = dsh_utils.CompileExpr(option.Assigns[j].Mapping)
				if err != nil {
					return errW(err, "project manifest invalid",
						reason("value invalid"),
						kv("path", m.manifestPath),
						kv("field", fmt.Sprintf("option.items[%d].assigns[%d].mapping", i, j)),
						kv("value", option.Assigns[j].Mapping),
					)
				}
			}
		}
	}
	for i := 0; i < len(m.Option.Verifies); i++ {
		if m.Option.Verifies[i] == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
			)
		}
		verify, err := dsh_utils.CompileExpr(m.Option.Verifies[i])
		if err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
				kv("value", m.Option.Verifies[i]),
			)
		}
		m.Option.verifies = append(m.Option.verifies, verify)
	}

	for i := 0; i < len(m.Script.Sources); i++ {
		src := m.Script.Sources[i]
		if err = src.init(m, "script", i); err != nil {
			return err
		}
	}
	for i := 0; i < len(m.Script.Imports); i++ {
		imp := m.Script.Imports[i]
		if err = imp.init(m, "script", i); err != nil {
			return err
		}
	}

	for i := 0; i < len(m.Config.Sources); i++ {
		src := m.Config.Sources[i]
		if err = src.init(m, "config", i); err != nil {
			return err
		}
	}
	for i := 0; i < len(m.Config.Imports); i++ {
		imp := m.Config.Imports[i]
		if err = imp.init(m, "config", i); err != nil {
			return err
		}
	}
	return nil
}

func (s *projectManifestSource) init(manifest *projectManifest, scope string, index int) (err error) {
	if s.Dir == "" {
		return errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("%s.sources[%d].dir", scope, index)),
		)
	}
	if s.Match != "" {
		s.match, err = dsh_utils.CompileExpr(s.Match)
		if err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.sources[%d].match", scope, index)),
				kv("value", s.Match),
			)
		}
	}
	return nil
}

func (i *projectManifestImport) init(manifest *projectManifest, scope string, index int) (err error) {
	importMethodCount := 0
	if i.Registry != nil {
		importMethodCount++
	}
	if i.Local != nil {
		importMethodCount++
	}
	if i.Git != nil {
		importMethodCount++
	}
	if importMethodCount != 1 {
		return errN("project manifest invalid",
			reason("[registry, local, git] must have only one"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("%s.imports[%d]", scope, index)),
		)
	} else if i.Registry != nil {
		if i.Registry.Name == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].registry.name", scope, index)),
			)
		}
		if i.Registry.Path == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].registry.path", scope, index)),
			)
		}
	} else if i.Local != nil {
		if i.Local.Dir == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].local.dir", scope, index)),
			)
		}
	} else if i.Git != nil {
		if i.Git.Url == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].git.url", scope, index)),
			)
		}
		if i.Git.Ref == "" {
			i.Git.Ref = "main"
		}
		if i.Git.url, err = url.Parse(i.Git.Url); err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].git.url", scope, index)),
				kv("value", i.Git.Url),
			)
		}
		i.Git.ref = parseGitRef(i.Git.Ref)
	}
	if i.Match != "" {
		i.match, err = dsh_utils.CompileExpr(i.Match)
		if err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].match", scope, index)),
				kv("value", i.Match),
			)
		}
	}
	return nil
}

func (i *projectManifestOptionItem) parseValue(rawValue string) (any, error) {
	if len(i.Choices) > 0 && !slices.Contains(i.Choices, rawValue) {
		return nil, errN("option parse value error",
			reason("not in choices"),
			kv("name", i.Name),
			kv("value", rawValue),
			kv("choices", i.Choices),
		)
	}
	var parsedValue any = nil
	switch i.Type {
	case projectManifestOptionItemTypeString:
		parsedValue = rawValue
	case projectManifestOptionItemTypeBool:
		parsedValue = rawValue == "true"
	case projectManifestOptionItemTypeInteger:
		integer, err := dsh_utils.ParseInteger(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse integer error"),
				kv("name", i.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = integer
	case projectManifestOptionItemTypeDecimal:
		decimal, err := dsh_utils.ParseDecimal(rawValue)
		if err != nil {
			return nil, errW(err, "option parse value error",
				reason("parse decimal error"),
				kv("name", i.Name),
				kv("value", rawValue),
			)
		}
		parsedValue = decimal
	default:
		// impossible
		panic(desc("option type not supported",
			kv("optionName", i.Name),
			kv("optionType", i.Type),
		))
	}
	return parsedValue, nil
}
