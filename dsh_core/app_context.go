package dsh_core

import (
	"dsh/dsh_utils"
)

type appContext struct {
	workspace      *Workspace
	logger         *dsh_utils.Logger
	Profile        *AppProfile
	Option         *appOption
	projectsByName map[string]*project
}

func newAppContext(workspace *Workspace, profile *AppProfile, option *appOption) *appContext {
	return &appContext{
		workspace:      workspace,
		logger:         workspace.logger,
		Profile:        profile,
		Option:         option,
		projectsByName: make(map[string]*project),
	}
}

func (c *appContext) loadProject(manifest *projectManifest) (p *project, err error) {
	if p, exist := c.projectsByName[manifest.Name]; exist {
		return p, nil
	}
	if p, err = loadProject(c, manifest); err != nil {
		return nil, err
	}
	c.projectsByName[manifest.Name] = p
	return p, nil
}
