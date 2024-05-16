package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"slices"
	"time"
)

type workspaceManifest struct {
	Clean         *workspaceManifestClean
	Shell         map[string]*workspaceManifestShell
	Registry      *workspaceManifestRegistry
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

type workspaceManifestRegistry struct {
	Scopes []*workspaceManifestRegistryScope
	Alters []*workspaceManifestRegistryAlter
	alters []*workspaceManifestRegistryAlter
}

type workspaceManifestRegistryScope struct {
	Name  string
	Local *workspaceManifestRegistryLocal
	Git   *workspaceManifestRegistryGit
}

type workspaceManifestRegistryAlter struct {
	Prefix string
	Local  *workspaceManifestRegistryLocal
	Git    *workspaceManifestRegistryGit
}

type workspaceManifestRegistryLocal struct {
	Dir string
}

type workspaceManifestRegistryGit struct {
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

func loadWorkspaceManifest(workspacePath string) (manifest *workspaceManifest, err error) {
	manifest = &workspaceManifest{
		Clean: &workspaceManifestClean{
			Output: &workspaceManifestCleanOutput{},
		},
		Shell:    make(map[string]*workspaceManifestShell),
		Registry: &workspaceManifestRegistry{},
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

	scopeNamesDict := make(map[string]bool)
	for i := 0; i < len(m.Registry.Scopes); i++ {
		scope := m.Registry.Scopes[i]
		if scope.Name == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("registry.scopes[%d].name", i)),
			)
		}
		if _, exist := scopeNamesDict[scope.Name]; exist {
			return errN("workspace manifest invalid",
				reason("value duplicate"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("registry.scopes[%d].name", i)),
				kv("value", scope.Name),
			)
		}
		scopeNamesDict[scope.Name] = true
		if scope.Local == nil && scope.Git == nil {
			return errN("workspace manifest invalid",
				reason("local and git are both nil"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("registry.scopes[%d]", i)),
			)
		} else if scope.Local != nil && scope.Git != nil {
			return errN("workspace manifest invalid",
				reason("local and git are both not nil"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("registry.scopes[%d]", i)),
			)
		} else if scope.Local != nil {
			if scope.Local.Dir == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("registry.scopes[%d].local.dir", i)),
				)
			}
		} else if scope.Git != nil {
			if scope.Git.Url == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("registry.scopes[%d].git.url", i)),
				)
			}
			if scope.Git.Ref == "" {
				scope.Git.Ref = "main"
			}
		}
	}

	alterPrefixesDict := make(map[string]bool)
	for i := 0; i < len(m.Registry.Alters); i++ {
		alter := m.Registry.Alters[i]
		if alter.Prefix == "" {
			return errN("workspace manifest invalid",
				reason("value empty"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("registry.alters[%d].prefix", i)),
			)
		}
		if _, exist := alterPrefixesDict[alter.Prefix]; exist {
			return errN("workspace manifest invalid",
				reason("value duplicate"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("registry.alters[%d].prefix", i)),
				kv("value", alter.Prefix),
			)
		}
		alterPrefixesDict[alter.Prefix] = true
		if alter.Local == nil && alter.Git == nil {
			return errN("workspace manifest invalid",
				reason("local and git are both nil"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("registry.alters[%d]", i)),
			)
		} else if alter.Local != nil && alter.Git != nil {
			return errN("workspace manifest invalid",
				reason("local and git are both not nil"),
				kv("path", m.manifestPath),
				kv("field", fmt.Sprintf("registry.alters[%d]", i)),
			)
		} else if alter.Local != nil {
			if alter.Local.Dir == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("registry.alters[%d].local.dir", i)),
				)
			}
		} else if alter.Git != nil {
			if alter.Git.Url == "" {
				return errN("workspace manifest invalid",
					reason("value empty"),
					kv("path", m.manifestPath),
					kv("field", fmt.Sprintf("registry.alters[%d].git.url", i)),
				)
			}
		}
	}
	if len(m.Registry.Alters) > 0 {
		m.Registry.alters = make([]*workspaceManifestRegistryAlter, len(m.Registry.Alters))
		copy(m.Registry.alters, m.Registry.Alters)
		slices.SortStableFunc(m.Registry.alters, func(l, r *workspaceManifestRegistryAlter) int {
			return len(r.Prefix) - len(l.Prefix)
		})
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
