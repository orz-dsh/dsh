package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
	"slices"
	"strings"
	"time"
)

// region manifest

type workspaceManifest struct {
	Profile       *workspaceManifestProfile
	Clean         *workspaceManifestClean
	Shell         workspaceManifestShell
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
		Shell:  workspaceManifestShell{},
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

type workspaceManifestShell map[string]*workspaceManifestShellItem

type workspaceManifestShellItem struct {
	Path string
	Exts []string
	Args []string
}

var workspaceDefaultShellExts = map[string][]string{
	"cmd":        {".cmd", ".bat"},
	"pwsh":       {".ps1"},
	"powershell": {".ps1"},
}

var workspaceDefaultShellExtsFallback = []string{".sh"}

var workspaceDefaultShellArgs = map[string][]string{
	"cmd":        {"/C", "{{.target.path}}"},
	"pwsh":       {"-NoProfile", "-File", "{{.target.path}}"},
	"powershell": {"-NoProfile", "-File", "{{.target.path}}"},
}

func (s workspaceManifestShell) init(manifest *workspaceManifest) (err error) {
	for k, v := range s {
		if v.Path != "" && !dsh_utils.IsFileExists(v.Path) {
			return errN("workspace manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("shell.%s.path", k)),
				kv("value", v.Path),
			)
		}
		for i := 0; i < len(v.Exts); i++ {
			if v.Exts[i] == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("shell.%s.exts[%d]", k, i)),
				)
			}
		}
		for i := 0; i < len(v.Args); i++ {
			if v.Args[i] == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("shell.%s.args[%d]", k, i)),
				)
			}
		}
	}

	return nil
}

func (s workspaceManifestShell) getPath(shell string) string {
	if i, exist := s[shell]; exist {
		return i.Path
	}
	return ""
}

func (s workspaceManifestShell) getExts(shell string) []string {
	if i, exist := s[shell]; exist {
		if i.Exts != nil {
			return i.Exts
		}
	}
	if exts, exist := workspaceDefaultShellExts[shell]; exist {
		if exts != nil {
			return exts
		}
	}
	return workspaceDefaultShellExtsFallback
}

func (s workspaceManifestShell) getArgs(shell string) []string {
	if i, ok := s[shell]; ok {
		if i.Args != nil {
			return i.Args
		}
	}
	if args, exist := workspaceDefaultShellArgs[shell]; exist {
		if args != nil {
			return args
		}
	}
	return nil
}

// endregion

// region import

type workspaceManifestImport struct {
	Registries          []*workspaceManifestImportRegistry
	Redirects           []*workspaceManifestImportRedirect
	registryDefinitions map[string]*importRegistryDefinition
	redirectDefinitions []*importRedirectDefinition
}

type workspaceManifestImportRegistry struct {
	Name  string
	Local *workspaceManifestImportLocal
	Git   *workspaceManifestImportGit
}

type workspaceManifestImportRedirect struct {
	Prefix string
	Local  *workspaceManifestImportLocal
	Git    *workspaceManifestImportGit
}

type workspaceManifestImportLocal struct {
	Dir string
}

type workspaceManifestImportGit struct {
	Url string
	Ref string
}

var workspaceDefaultImportRegistryDefinitions = map[string]*importRegistryDefinition{
	"orz-dsh": {
		Name: "orz-dsh",
		Git: &importGitDefinition{
			Url: "https://github.com/orz-dsh/{{.path}}.git",
			Ref: "main",
		},
	},
	"orz-ops": {
		Name: "orz-ops",
		Git: &importGitDefinition{
			Url: "https://github.com/orz-ops/{{.path}}.git",
			Ref: "main",
		},
	},
}

func (imp *workspaceManifestImport) init(manifest *workspaceManifest) (err error) {
	registryDefinitions := make(map[string]*importRegistryDefinition)
	for i := 0; i < len(imp.Registries); i++ {
		registry := imp.Registries[i]
		if registry.Name == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.registries[%d].name", i)),
			)
		}
		if _, exist := registryDefinitions[registry.Name]; exist {
			return errN("workspace manifest invalid",
				reason("value duplicate"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.registries[%d].name", i)),
				kv("value", registry.Name),
			)
		}
		if err = imp.checkImportMode(manifest, registry.Local, registry.Git, "registries", i); err != nil {
			return err
		}
		if registry.Git != nil {
			if registry.Git.Ref == "" {
				registry.Git.Ref = "main"
			}
		}
		if registry.Local != nil {
			registryDefinitions[registry.Name] = newImportRegistryLocalDefinition(registry.Name, registry.Local.Dir)
		} else if registry.Git != nil {
			registryDefinitions[registry.Name] = newImportRegistryGitDefinition(registry.Name, registry.Git.Url, registry.Git.Ref)
		} else {
			impossible()
		}
	}
	imp.registryDefinitions = registryDefinitions

	redirectPrefixes := make(map[string]bool)
	var redirectDefinitions []*importRedirectDefinition
	for i := 0; i < len(imp.Redirects); i++ {
		redirect := imp.Redirects[i]
		if redirect.Prefix == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].prefix", i)),
			)
		}
		if _, exist := redirectPrefixes[redirect.Prefix]; exist {
			return errN("workspace manifest invalid",
				reason("value duplicate"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].prefix", i)),
				kv("value", redirect.Prefix),
			)
		}
		redirectPrefixes[redirect.Prefix] = true
		if err = imp.checkImportMode(manifest, redirect.Local, redirect.Git, "redirects", i); err != nil {
			return err
		}
		if redirect.Local != nil {
			redirectDefinitions = append(redirectDefinitions, newImportRedirectLocalDefinition(redirect.Prefix, redirect.Local.Dir))
		} else if redirect.Git != nil {
			redirectDefinitions = append(redirectDefinitions, newImportRedirectGitDefinition(redirect.Prefix, redirect.Git.Url, redirect.Git.Ref))
		} else {
			impossible()
		}
	}
	if len(redirectDefinitions) > 0 {
		slices.SortStableFunc(redirectDefinitions, func(l, r *importRedirectDefinition) int {
			return len(r.Prefix) - len(l.Prefix)
		})
		imp.redirectDefinitions = redirectDefinitions
	}

	return nil
}

func (imp *workspaceManifestImport) checkImportMode(manifest *workspaceManifest, local *workspaceManifestImportLocal, git *workspaceManifestImportGit, scope string, index int) error {
	importModeCount := 0
	if local != nil {
		importModeCount++
	}
	if git != nil {
		importModeCount++
	}
	if importModeCount != 1 {
		return errN("workspace manifest invalid",
			reason("[local, git] must have only one"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("import.%s[%d]", scope, index)),
		)
	} else if local != nil {
		if local.Dir == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.%s[%d].local.dir", scope, index)),
			)
		}
	} else {
		if git.Url == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.%s[%d].git.url", scope, index)),
			)
		}
	}
	return nil
}

func (imp *workspaceManifestImport) getRegistryDefinition(name string) *importRegistryDefinition {
	if definition, exist := imp.registryDefinitions[name]; exist {
		return definition
	}
	if definition, exist := workspaceDefaultImportRegistryDefinitions[name]; exist {
		return definition
	}
	return nil
}

func (imp *workspaceManifestImport) getRedirectDefinition(resources []string) (*importRedirectDefinition, string) {
	for i := 0; i < len(resources); i++ {
		resource := resources[i]
		for j := 0; j < len(imp.redirectDefinitions); j++ {
			definition := imp.redirectDefinitions[j]
			if path, found := strings.CutPrefix(resource, definition.Prefix); found {
				return definition, path
			}
		}
	}
	return nil, ""
}

// endregion
