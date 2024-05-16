package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"time"
)

type workspaceManifest struct {
	Clean         *workspaceManifestClean
	Shell         map[string]*workspaceManifestShell
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
		Shell: make(map[string]*workspaceManifestShell),
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
