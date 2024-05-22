package dsh_core

import (
	"dsh/dsh_utils"
)

type appContext struct {
	logger         *dsh_utils.Logger
	workspace      *Workspace
	manifest       *projectManifest
	Profile        *AppProfile
	Option         *appOption
	projectsByName map[string]*project
}

func newAppContext(workspace *Workspace, manifest *projectManifest, profile *AppProfile, option *appOption) *appContext {
	return &appContext{
		logger:         workspace.logger,
		workspace:      workspace,
		manifest:       manifest,
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

func (c *appContext) loadMainProject() (p *project, err error) {
	return c.loadProject(c.manifest)
}

func (c *appContext) isMainProject(manifest *projectManifest) bool {
	return c.manifest.Name == manifest.Name
}
