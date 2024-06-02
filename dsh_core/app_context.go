package dsh_core

type appContext struct {
	logger         *Logger
	workspace      *Workspace
	evaluator      *Evaluator
	profile        *appProfile
	option         *appOption
	projectsByName map[string]*appProject
}

func newAppContext(workspace *Workspace, evaluator *Evaluator, profile *appProfile, option *appOption) *appContext {
	return &appContext{
		logger:         workspace.logger,
		workspace:      workspace,
		evaluator:      evaluator,
		profile:        profile,
		option:         option,
		projectsByName: map[string]*appProject{},
	}
}

func (c *appContext) loadProject(projectEntity *projectSchema) (project *appProject, err error) {
	if existProject, exist := c.projectsByName[projectEntity.Name]; exist {
		return existProject, nil
	}
	if project, err = makeAppProject(c, projectEntity); err != nil {
		return nil, err
	}
	c.projectsByName[projectEntity.Name] = project
	return project, nil
}

func (c *appContext) loadProjectByTarget(target *projectLinkTarget) (project *appProject, err error) {
	entity, err := c.profile.getProjectEntityByLinkTarget(target)
	if err != nil {
		return nil, errW(err, "load project error",
			kv("reason", "load project entity error"),
			kv("target", target),
		)
	}
	project, err = c.loadProject(entity)
	if err != nil {
		return nil, errW(err, "load project error",
			kv("target", target),
		)
	}
	return project, nil
}
