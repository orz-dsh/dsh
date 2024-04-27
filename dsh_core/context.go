package dsh_core

import "dsh/dsh_utils"

type Context struct {
	Workspace      *Workspace
	Logger         *dsh_utils.Logger
	OptionSelector *OptionSelector
	Project        *Project
}

type OptionSelector struct {
}

func NewContext(workspace *Workspace, logger *dsh_utils.Logger) *Context {
	return &Context{
		Workspace:      workspace,
		Logger:         logger,
		OptionSelector: &OptionSelector{},
	}
}
