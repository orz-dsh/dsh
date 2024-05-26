package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
	"regexp"
	"time"
)

// region manifest

type workspaceManifest struct {
	Profile       *workspaceManifestProfile
	Clean         *workspaceManifestClean
	Shell         *workspaceManifestShell
	Import        *workspaceManifestImport
	manifestPath  string
	manifestType  manifestMetadataType
	workspacePath string
}

func loadWorkspaceManifest(workspacePath string) (manifest *workspaceManifest, err error) {
	manifest = &workspaceManifest{
		Profile: &workspaceManifestProfile{},
		Clean: &workspaceManifestClean{
			Output: &workspaceManifestCleanOutput{},
		},
		Shell:  &workspaceManifestShell{},
		Import: &workspaceManifestImport{},
	}
	metadata, err := loadManifestFromDir(workspacePath, []string{"workspace"}, manifest, false)
	if err != nil {
		return nil, errW(err, "load workspace manifest error",
			reason("load manifest from dir error"),
			kv("workspacePath", workspacePath),
		)
	}
	if metadata != nil {
		manifest.manifestPath = metadata.ManifestPath
		manifest.manifestType = metadata.ManifestType
	}
	manifest.workspacePath = workspacePath
	if err = manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *workspaceManifest) DescExtraKeyValues() KVS {
	return KVS{
		kv("manifestPath", m.manifestPath),
		kv("manifestType", m.manifestType),
		kv("workspacePath", m.workspacePath),
	}
}

func (m *workspaceManifest) init() (err error) {
	if err = m.Profile.init(m); err != nil {
		return err
	}
	if err = m.Clean.init(m); err != nil {
		return err
	}
	if err = m.Shell.init(m); err != nil {
		return err
	}
	if err = m.Import.init(m); err != nil {
		return err
	}

	return nil
}

// endregion

// region profile

type workspaceManifestProfile struct {
	Items []*workspaceManifestProfileItem
}

type workspaceManifestProfileItem struct {
	File     string
	Optional bool
	Match    string
	match    *vm.Program
}

func (p *workspaceManifestProfile) init(manifest *workspaceManifest) (err error) {
	for i := 0; i < len(p.Items); i++ {
		item := p.Items[i]
		if item.File == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("profile.items[%d].file", i)),
			)
		}
		if item.Match != "" {
			item.match, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "workspace manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("profile.items[%d].match", i)),
					kv("value", item.Match),
				)
			}
		}
	}
	return nil
}

// endregion

// region clean

type workspaceManifestClean struct {
	Output *workspaceManifestCleanOutput
}

type workspaceManifestCleanOutput struct {
	Count   *int
	Expires string
	count   int
	expires time.Duration
}

const workspaceDefaultCleanOutputCount = 3
const workspaceDefaultCleanOutputExpires = 24 * time.Hour

func (c *workspaceManifestClean) init(manifest *workspaceManifest) (err error) {
	if c.Output.Expires != "" {
		c.Output.expires, err = time.ParseDuration(c.Output.Expires)
		if err != nil {
			return errN("workspace manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", "clean.output.expires"),
				kv("value", c.Output.Expires),
			)
		}
	} else {
		c.Output.expires = workspaceDefaultCleanOutputExpires
	}

	if c.Output.Count != nil {
		value := *c.Output.Count
		if value <= 0 {
			return errN("workspace manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", "clean.output.count"),
				kv("value", value),
			)
		}
		c.Output.count = value
	} else {
		c.Output.count = workspaceDefaultCleanOutputCount
	}

	return nil
}

// endregion

// region shell

type workspaceManifestShell struct {
	Items       []*workspaceManifestShellItem
	definitions workspaceShellDefinitions
}

type workspaceManifestShellItem struct {
	Name  string
	Path  string
	Exts  []string
	Args  []string
	Match string
}

func (s *workspaceManifestShell) init(manifest *workspaceManifest) (err error) {
	definitions := workspaceShellDefinitions{}
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if item.Name == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("shell.items[%d].name", i)),
			)
		}
		if item.Path != "" && !dsh_utils.IsFileExists(item.Path) {
			return errN("workspace manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("shell.items[%d].path", i)),
				kv("value", item.Path),
			)
		}
		for j := 0; j < len(item.Exts); j++ {
			if item.Exts[j] == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("shell.items[%d].exts[%d]", i, j)),
				)
			}
		}
		for j := 0; j < len(item.Args); j++ {
			if item.Args[j] == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("shell.items[%d].args[%d]", i, j)),
				)
			}
		}
		var matchExpr *vm.Program
		if item.Match != "" {
			matchExpr, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "app profile manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].match", i)),
					kv("value", item.Match),
				)
			}
		}
		definitions[item.Name] = append(definitions[item.Name], newWorkspaceShellDefinition(item.Name, item.Path, item.Exts, item.Args, item.Match, matchExpr))
	}

	s.definitions = definitions
	return nil
}

// endregion

// region import

type workspaceManifestImport struct {
	Registries          []*workspaceManifestImportRegistry
	Redirects           []*workspaceManifestImportRedirect
	registryDefinitions workspaceImportRegistryDefinitions
	redirectDefinitions workspaceImportRedirectDefinitions
}

type workspaceManifestImportRegistry struct {
	Name  string
	Link  string
	Match string
}

type workspaceManifestImportRedirect struct {
	Regex string
	Link  string
	Match string
}

func (imp *workspaceManifestImport) init(manifest *workspaceManifest) (err error) {
	registryDefinitions := workspaceImportRegistryDefinitions{}
	for i := 0; i < len(imp.Registries); i++ {
		registry := imp.Registries[i]
		if registry.Name == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.registries[%d].name", i)),
			)
		}

		if registry.Link == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.registries[%d].link", i)),
			)
		}
		// TODO: check link template

		var matchExpr *vm.Program
		if registry.Match != "" {
			matchExpr, err = dsh_utils.CompileExpr(registry.Match)
			if err != nil {
				return errW(err, "workspace manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("import.registries[%d].match", i)),
					kv("value", registry.Match),
				)
			}
		}
		registryDefinitions[registry.Name] = append(registryDefinitions[registry.Name], newWorkspaceImportRegistryDefinition(registry.Name, registry.Link, registry.Match, matchExpr))
	}

	redirectDefinitions := workspaceImportRedirectDefinitions{}
	for i := 0; i < len(imp.Redirects); i++ {
		redirect := imp.Redirects[i]
		if redirect.Regex == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].regex", i)),
			)
		}
		regex, err := regexp.Compile(redirect.Regex)
		if err != nil {
			return errW(err, "workspace manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].regex", i)),
				kv("value", redirect.Regex),
			)
		}

		if redirect.Link == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].link", i)),
			)
		}
		// TODO: check link template

		var matchExpr *vm.Program
		if redirect.Match != "" {
			matchExpr, err = dsh_utils.CompileExpr(redirect.Match)
			if err != nil {
				return errW(err, "workspace manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("import.redirects[%d].match", i)),
					kv("value", redirect.Match),
				)
			}
		}
		redirectDefinitions = append(redirectDefinitions, newWorkspaceImportRedirectDefinition(regex, redirect.Link, redirect.Match, matchExpr))
	}

	imp.registryDefinitions = registryDefinitions
	imp.redirectDefinitions = redirectDefinitions
	return nil
}

// endregion
