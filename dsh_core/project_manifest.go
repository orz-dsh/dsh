package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
	"net/url"
	"regexp"
	"slices"
)

// region manifest

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

func loadProjectManifest(projectPath string) (manifest *projectManifest, err error) {
	manifest = &projectManifest{
		Runtime: &projectManifestRuntime{},
		Option:  &projectManifestOption{},
		Script:  &projectManifestScript{},
		Config:  &projectManifestConfig{},
	}
	metadata, err := loadManifestFromDir(projectPath, []string{"project"}, manifest, true)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("load manifest from dir error"),
			kv("projectPath", projectPath),
		)
	}
	manifest.manifestPath = metadata.ManifestPath
	manifest.manifestType = metadata.ManifestType
	manifest.projectPath = projectPath
	if err = manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *projectManifest) DescExtraKeyValues() KVS {
	return KVS{
		kv("projectPath", m.projectPath),
		kv("manifestPath", m.manifestPath),
		kv("manifestType", m.manifestType),
	}
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
	if err = m.Runtime.init(m); err != nil {
		return err
	}
	if err = m.Option.init(m); err != nil {
		return err
	}
	if err = m.Script.init(m); err != nil {
		return err
	}
	if err = m.Config.init(m); err != nil {
		return err
	}
	return nil
}

// endregion

// region runtime

type projectManifestRuntime struct {
	MinVersion dsh_utils.Version `yaml:"minVersion" toml:"minVersion" json:"minVersion"`
	MaxVersion dsh_utils.Version `yaml:"maxVersion" toml:"maxVersion" json:"maxVersion"`
}

func (r *projectManifestRuntime) init(manifest *projectManifest) (err error) {
	if err = dsh_utils.CheckRuntimeVersion(r.MinVersion, r.MaxVersion); err != nil {
		return errW(err, "project manifest invalid",
			reason("runtime incompatible"),
			kv("path", manifest.manifestPath),
			kv("field", "runtime"),
			kv("minVersion", r.MinVersion),
			kv("maxVersion", r.MaxVersion),
			kv("runtimeVersion", dsh_utils.GetRuntimeVersion()),
		)
	}
	return nil
}

// endregion

// region option

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

type projectManifestOptionItemType string

const (
	projectManifestOptionItemTypeString  projectManifestOptionItemType = "string"
	projectManifestOptionItemTypeBool    projectManifestOptionItemType = "bool"
	projectManifestOptionItemTypeInteger projectManifestOptionItemType = "integer"
	projectManifestOptionItemTypeDecimal projectManifestOptionItemType = "decimal"
)

func (o *projectManifestOption) init(manifest *projectManifest) (err error) {
	optionNamesDict := make(map[string]bool)
	for i := 0; i < len(o.Items); i++ {
		option := o.Items[i]
		if option.Name == "" {
			return errN("project manifest invalid",
				reason("name empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].name", i)),
			)
		}
		if checked := projectNameCheckRegex.MatchString(option.Name); !checked {
			return errN("project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.items[%d].name", i)),
				kv("value", option.Name),
			)
		}
		if _, exist := optionNamesDict[option.Name]; exist {
			return errN("project manifest invalid",
				reason("name duplicated"),
				kv("path", manifest.manifestPath),
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
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].type", i)),
					kv("value", option.Type),
				)
			}

			defaultRawValue := *option.Default
			defaultParsedValue, err := option.parseValue(defaultRawValue)
			if err != nil {
				return errW(err, "project manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
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
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", i, j)),
				)
			}
			if option.Assigns[j].Project == manifest.Name {
				return errN("project manifest invalid",
					reason("can not assign to self project option"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d].project", i, j)),
				)
			}
			if option.Assigns[j].Option == "" {
				return errN("project manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("option.items[%d].assigns[%d].option", i, j)),
				)
			}
			assignTarget := option.Assigns[j].Project + "." + option.Assigns[j].Option
			if _, exists := assignTargetsDict[assignTarget]; exists {
				return errN("project manifest invalid",
					reason("option assign target duplicated"),
					kv("path", manifest.manifestPath),
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
						kv("path", manifest.manifestPath),
						kv("field", fmt.Sprintf("option.items[%d].assigns[%d].mapping", i, j)),
						kv("value", option.Assigns[j].Mapping),
					)
				}
			}
		}
	}
	for i := 0; i < len(o.Verifies); i++ {
		if o.Verifies[i] == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
			)
		}
		verify, err := dsh_utils.CompileExpr(o.Verifies[i])
		if err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("option.verifies[%d]", i)),
				kv("value", o.Verifies[i]),
			)
		}
		o.verifies = append(o.verifies, verify)
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
		impossible()
	}
	return parsedValue, nil
}

// endregion

// region script

type projectManifestScript struct {
	Sources []*projectManifestSource
	Imports []*projectManifestImport
}

func (s *projectManifestScript) init(manifest *projectManifest) (err error) {
	for i := 0; i < len(s.Sources); i++ {
		src := s.Sources[i]
		if err = src.init(manifest, "script", i); err != nil {
			return err
		}
	}
	for i := 0; i < len(s.Imports); i++ {
		imp := s.Imports[i]
		if err = imp.init(manifest, "script", i); err != nil {
			return err
		}
	}
	return nil
}

// endregion

// region config

type projectManifestConfig struct {
	Sources []*projectManifestSource
	Imports []*projectManifestImport
}

func (c *projectManifestConfig) init(manifest *projectManifest) (err error) {
	for i := 0; i < len(c.Sources); i++ {
		src := c.Sources[i]
		if err = src.init(manifest, "config", i); err != nil {
			return err
		}
	}
	for i := 0; i < len(c.Imports); i++ {
		imp := c.Imports[i]
		if err = imp.init(manifest, "config", i); err != nil {
			return err
		}
	}
	return nil
}

// endregion

// region source

type projectManifestSource struct {
	Dir   string
	Files []string
	Match string
	match *vm.Program
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

// endregion

// region import

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

func (i *projectManifestImport) init(manifest *projectManifest, scope string, index int) (err error) {
	importModeCount := 0
	if i.Registry != nil {
		importModeCount++
	}
	if i.Local != nil {
		importModeCount++
	}
	if i.Git != nil {
		importModeCount++
	}
	if importModeCount != 1 {
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

// endregion
