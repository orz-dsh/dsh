package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
)

type workspaceManifest struct {
	Shell         map[string]*workspaceManifestShell
	manifestPath  string
	manifestType  manifestMetadataType
	workspacePath string
}

type workspaceManifestShell struct {
	Path string
	Exts []string
	Args []string
}

func loadWorkspaceManifest(workspacePath string) (m *workspaceManifest, err error) {
	m = &workspaceManifest{
		Shell: make(map[string]*workspaceManifestShell),
	}
	metadata, err := loadManifest(workspacePath, []string{"workspace"}, m, false)
	if err != nil {
		return nil, errW(err, "load workspace manifest error",
			reason("load manifest error"),
			kv("workspacePath", workspacePath),
		)
	}
	if metadata != nil {
		m.manifestPath = metadata.manifestPath
		m.manifestType = metadata.manifestType
	}
	m.workspacePath = workspacePath
	if err = m.init(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *workspaceManifest) init() (err error) {
	if m.manifestPath == "" {
		return nil
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
	if s, ok := m.Shell[shell]; ok {
		return s.Path
	}
	return ""
}

func (m *workspaceManifest) getShellExts(shell string) []string {
	if s, ok := m.Shell[shell]; ok {
		return s.Exts
	}
	return nil
}

func (m *workspaceManifest) getShellArgs(shell string) []string {
	if s, ok := m.Shell[shell]; ok {
		return s.Args
	}
	return nil
}
