package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"slices"
	"strings"
	"time"
)

type workspaceManifest struct {
	Clean         *workspaceManifestClean
	Shell         map[string]*workspaceManifestShell
	Import        *workspaceManifestImport
	manifestPath  string
	manifestType  manifestMetadataType
	workspacePath string
}

type workspaceManifestClean struct {
	Output *workspaceManifestCleanOutput
}

type workspaceManifestCleanOutput struct {
	Count   *int
	Expires string
	count   int
	expires time.Duration
}

type workspaceManifestShell struct {
	Path string
	Exts []string
	Args []string
}

type workspaceManifestImport struct {
	Registries       []*workspaceManifestImportRegistry
	Redirects        []*workspaceManifestImportRedirect
	registriesByName map[string]*workspaceManifestImportRegistry
	redirectsSorted  []*workspaceManifestImportRedirect
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

const workspaceDefaultCleanOutputCount = 3
const workspaceDefaultCleanOutputExpires = 24 * time.Hour

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

var workspaceDefaultImportRegistries = map[string]*workspaceManifestImportRegistry{
	"orz-dsh": {
		Name: "orz-dsh",
		Git: &workspaceManifestImportGit{
			Url: "https://github.com/orz-dsh/{{.path}}.git",
			Ref: "main",
		},
	},
	"orz-ops": {
		Name: "orz-ops",
		Git: &workspaceManifestImportGit{
			Url: "https://github.com/orz-ops/{{.path}}.git",
			Ref: "main",
		},
	},
}

func loadWorkspaceManifest(workspacePath string) (manifest *workspaceManifest, err error) {
	manifest = &workspaceManifest{
		Clean: &workspaceManifestClean{
			Output: &workspaceManifestCleanOutput{},
		},
		Shell:  make(map[string]*workspaceManifestShell),
		Import: &workspaceManifestImport{},
	}
	metadata, err := loadManifest(workspacePath, []string{"workspace"}, manifest, false)
	if err != nil {
		return nil, errW(err, "load workspace manifest error",
			reason("load manifest error"),
			kv("workspacePath", workspacePath),
		)
	}
	if metadata != nil {
		manifest.manifestPath = metadata.manifestPath
		manifest.manifestType = metadata.manifestType
	}
	manifest.workspacePath = workspacePath
	if err = manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *workspaceManifest) init() (err error) {
	if m.Clean.Output.Expires != "" {
		m.Clean.Output.expires, err = time.ParseDuration(m.Clean.Output.Expires)
		if err != nil {
			return errN("workspace manifest invalid",
				reason("value invalid"),
				kv("path", m.manifestPath),
				kv("field", "clean.output.expires"),
				kv("value", m.Clean.Output.Expires),
			)
		}
	} else {
		m.Clean.Output.expires = workspaceDefaultCleanOutputExpires
	}

	if m.Clean.Output.Count != nil {
		value := *m.Clean.Output.Count
		if value <= 0 {
			return errN("workspace manifest invalid",
				reason("value invalid"),
				kv("path", m.manifestPath),
				kv("field", "clean.output.count"),
				kv("value", value),
			)
		}
		m.Clean.Output.count = value
	} else {
		m.Clean.Output.count = workspaceDefaultCleanOutputCount
	}

	for k, v := range m.Shell {
		if v.Path != "" && !dsh_utils.IsFileExists(v.Path) {
			return errN("workspace manifest invalid",
				reason("value invalid"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("shell.%s.path", k)),
				kv("value", v.Path),
			)
		}
		for i := 0; i < len(v.Exts); i++ {
			if v.Exts[i] == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("shell.%s.exts[%d]", k, i)),
				)
			}
		}
		for i := 0; i < len(v.Args); i++ {
			if v.Args[i] == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("shell.%s.args[%d]", k, i)),
				)
			}
		}
	}

	registriesByName := make(map[string]*workspaceManifestImportRegistry)
	for i := 0; i < len(m.Import.Registries); i++ {
		registry := m.Import.Registries[i]
		if registry.Name == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("import.registries[%d].name", i)),
			)
		}
		if _, exist := registriesByName[registry.Name]; exist {
			return errN("workspace manifest invalid",
				reason("value duplicate"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("import.registries[%d].name", i)),
				kv("value", registry.Name),
			)
		}
		registriesByName[registry.Name] = registry
		if err = m.checkImportMethod(registry.Local, registry.Git, "registries", i); err != nil {
			return err
		}
		if registry.Git != nil {
			if registry.Git.Ref == "" {
				registry.Git.Ref = "main"
			}
		}
	}
	m.Import.registriesByName = registriesByName

	redirectPrefixesDict := make(map[string]bool)
	for i := 0; i < len(m.Import.Redirects); i++ {
		redirect := m.Import.Redirects[i]
		if redirect.Prefix == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].prefix", i)),
			)
		}
		if _, exist := redirectPrefixesDict[redirect.Prefix]; exist {
			return errN("workspace manifest invalid",
				reason("value duplicate"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].prefix", i)),
				kv("value", redirect.Prefix),
			)
		}
		redirectPrefixesDict[redirect.Prefix] = true
		if err = m.checkImportMethod(redirect.Local, redirect.Git, "redirects", i); err != nil {
			return err
		}
	}
	if len(m.Import.Redirects) > 0 {
		m.Import.redirectsSorted = make([]*workspaceManifestImportRedirect, len(m.Import.Redirects))
		copy(m.Import.redirectsSorted, m.Import.Redirects)
		slices.SortStableFunc(m.Import.redirectsSorted, func(l, r *workspaceManifestImportRedirect) int {
			return len(r.Prefix) - len(l.Prefix)
		})
	}

	return nil
}

func (m *workspaceManifest) checkImportMethod(local *workspaceManifestImportLocal, git *workspaceManifestImportGit, scope string, index int) error {
	importMethodCount := 0
	if local != nil {
		importMethodCount++
	}
	if git != nil {
		importMethodCount++
	}
	if importMethodCount != 1 {
		return errN("workspace manifest invalid",
			reason("[local, git] must have only one"),
			kv("path", m.manifestPath),
			kv("field", fmt.Sprintf("import.%s[%d]", scope, index)),
		)
	} else if local != nil {
		if local.Dir == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("import.%s[%d].local.dir", scope, index)),
			)
		}
	} else {
		if git.Url == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("import.%s[%d].git.url", scope, index)),
			)
		}
	}
	return nil
}

func (m *workspaceManifest) getShellPath(shell string) string {
	if s, exist := m.Shell[shell]; exist {
		return s.Path
	}
	return ""
}

func (m *workspaceManifest) getShellExts(shell string) []string {
	if s, exist := m.Shell[shell]; exist {
		if s.Exts != nil {
			return s.Exts
		}
	}
	if exts, exist := workspaceDefaultShellExts[shell]; exist {
		if exts != nil {
			return exts
		}
	}
	return workspaceDefaultShellExtsFallback
}

func (m *workspaceManifest) getShellArgs(shell string) []string {
	if s, ok := m.Shell[shell]; ok {
		if s.Args != nil {
			return s.Args
		}
	}
	if args, exist := workspaceDefaultShellArgs[shell]; exist {
		if args != nil {
			return args
		}
	}
	return nil
}

func (m *workspaceManifest) getImportRegistry(name string) *workspaceManifestImportRegistry {
	if registry, exist := m.Import.registriesByName[name]; exist {
		return registry
	}
	if registry, exist := workspaceDefaultImportRegistries[name]; exist {
		return registry
	}
	return nil
}

func (m *workspaceManifest) getImportRedirect(path string) *workspaceManifestImportRedirect {
	for i := 0; i < len(m.Import.redirectsSorted); i++ {
		redirect := m.Import.redirectsSorted[i]
		if strings.HasPrefix(path, redirect.Prefix) {
			return redirect
		}
	}
	return nil
}
