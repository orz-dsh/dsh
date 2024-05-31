package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
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
	metadata, err := dsh_utils.DeserializeByDir(workspacePath, []string{"workspace"}, manifest, false)
	if err != nil {
		return nil, errW(err, "load workspace manifest error",
			reason("load manifest from dir error"),
			kv("workspacePath", workspacePath),
		)
	}
	if metadata != nil {
		manifest.manifestPath = metadata.Path
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
	Items    []*workspaceManifestProfileItem
	entities workspaceProfileEntitySet
}

type workspaceManifestProfileItem struct {
	File     string
	Optional bool
	Match    string
}

func (p *workspaceManifestProfile) init(manifest *workspaceManifest) (err error) {
	entities := workspaceProfileEntitySet{}
	for i := 0; i < len(p.Items); i++ {
		item := p.Items[i]
		if item.File == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("profile.items[%d].file", i)),
			)
		}

		var matchObj *EvalExpr
		if item.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "workspace manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("profile.items[%d].match", i)),
					kv("value", item.Match),
				)
			}
		}
		entities = append(entities, newWorkspaceProfileEntity(item.File, item.Optional, item.Match, matchObj))
	}

	p.entities = entities
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
	Items    []*workspaceManifestShellItem
	entities workspaceShellEntitySet
}

type workspaceManifestShellItem struct {
	Name  string
	Path  string
	Exts  []string
	Args  []string
	Match string
}

func (s *workspaceManifestShell) init(manifest *workspaceManifest) (err error) {
	entities := workspaceShellEntitySet{}
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
		var matchObj *EvalExpr
		if item.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(item.Match)
			if err != nil {
				return errW(err, "app profile manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("workspace.shell.items[%d].match", i)),
					kv("value", item.Match),
				)
			}
		}
		entities[item.Name] = append(entities[item.Name], newWorkspaceShellEntity(item.Name, item.Path, item.Exts, item.Args, item.Match, matchObj))
	}

	s.entities = entities
	return nil
}

// endregion

// region import

type workspaceManifestImport struct {
	Registries       []*workspaceManifestImportRegistry
	Redirects        []*workspaceManifestImportRedirect
	registryEntities workspaceImportRegistryEntitySet
	redirectEntities workspaceImportRedirectEntitySet
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
	registryEntities := workspaceImportRegistryEntitySet{}
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

		var matchObj *EvalExpr
		if registry.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(registry.Match)
			if err != nil {
				return errW(err, "workspace manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("import.registries[%d].match", i)),
					kv("value", registry.Match),
				)
			}
		}
		registryEntities[registry.Name] = append(registryEntities[registry.Name], newWorkspaceImportRegistryEntity(registry.Name, registry.Link, registry.Match, matchObj))
	}

	redirectEntities := workspaceImportRedirectEntitySet{}
	for i := 0; i < len(imp.Redirects); i++ {
		redirect := imp.Redirects[i]
		if redirect.Regex == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].regex", i)),
			)
		}
		regexObj, err := regexp.Compile(redirect.Regex)
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

		var matchObj *EvalExpr
		if redirect.Match != "" {
			matchObj, err = dsh_utils.CompileExpr(redirect.Match)
			if err != nil {
				return errW(err, "workspace manifest invalid",
					reason("value invalid"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("import.redirects[%d].match", i)),
					kv("value", redirect.Match),
				)
			}
		}
		redirectEntities = append(redirectEntities, newWorkspaceImportRedirectEntity(redirect.Regex, redirect.Link, redirect.Match, regexObj, matchObj))
	}

	imp.registryEntities = registryEntities
	imp.redirectEntities = redirectEntities
	return nil
}

// endregion
