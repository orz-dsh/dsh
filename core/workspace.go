package core

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/internal"
	. "github.com/orz-dsh/dsh/utils"
)

type Workspace struct {
	core *WorkspaceCore
}

func NewWorkspace(environment *Environment, dir string) (workspace *Workspace, err error) {
	core, err := NewWorkspaceCore(environment.core, dir)
	if err != nil {
		return nil, err
	}
	workspace = &Workspace{core: core}
	return workspace, nil
}

func (w *Workspace) DescExtraKeyValues() KVS {
	return KVS{
		KV("core", w.core),
	}
}

func (w *Workspace) GetDir() string {
	return w.core.Dir
}

func (w *Workspace) Clean(options WorkspaceCleanOptions) error {
	return w.core.Clean(options)
}

func (w *Workspace) NewAppBuilder() *ApplicationBuilder {
	return newAppBuilder(w.core)
}
