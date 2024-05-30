package dsh_core

type appContext struct {
	systemInfo     *SystemInfo
	logger         *Logger
	workspace      *Workspace
	evaluator      *Evaluator
	manifest       *ProjectManifest
	profile        *appProfile
	option         *appOption
	projectsByName map[string]*Project
}

func newAppContext(workspace *Workspace, profile *appProfile, manifest *ProjectManifest, option *appOption) *appContext {
	return &appContext{
		systemInfo:     workspace.systemInfo,
		logger:         workspace.logger,
		workspace:      workspace,
		evaluator:      profile.evaluator,
		manifest:       manifest,
		profile:        profile,
		option:         option,
		projectsByName: map[string]*Project{},
	}
}

func (c *appContext) loadProject(manifest *ProjectManifest) (project *Project, err error) {
	if existProject, exist := c.projectsByName[manifest.projectName]; exist {
		return existProject, nil
	}
	if project, err = makeProject(c, manifest); err != nil {
		return nil, err
	}
	c.projectsByName[manifest.projectName] = project
	return project, nil
}

func (c *appContext) loadMainProject() (p *Project, err error) {
	return c.loadProject(c.manifest)
}

func (c *appContext) isMainProject(manifest *ProjectManifest) bool {
	return c.manifest.projectName == manifest.projectName
}
