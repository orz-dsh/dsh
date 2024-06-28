package core

type appContext struct {
	logger         *Logger
	workspace      *Workspace
	evaluator      *Evaluator
	profile        *appSetting
	option         *appOption
	projectsByName map[string]*projectEntity
}

func newAppContext(workspace *Workspace, evaluator *Evaluator, profile *appSetting, option *appOption) *appContext {
	return &appContext{
		logger:         workspace.logger,
		workspace:      workspace,
		evaluator:      evaluator,
		profile:        profile,
		option:         option,
		projectsByName: map[string]*projectEntity{},
	}
}

func (c *appContext) loadProject(setting *projectSetting) (project *projectEntity, err error) {
	if existProject, exist := c.projectsByName[setting.Name]; exist {
		return existProject, nil
	}
	if project, err = newProjectEntity(c, setting, nil); err != nil {
		return nil, err
	}
	c.projectsByName[setting.Name] = project
	return project, nil
}

func (c *appContext) loadProjectByTarget(target *projectLinkTarget) (project *projectEntity, err error) {
	setting, err := c.profile.getProjectSettingByLinkTarget(target)
	if err != nil {
		return nil, errW(err, "load project error",
			kv("reason", "load project setting error"),
			kv("target", target),
		)
	}
	project, err = c.loadProject(setting)
	if err != nil {
		return nil, errW(err, "load project error",
			kv("target", target),
		)
	}
	return project, nil
}

func (c *appContext) getProject(name string) *projectEntity {
	return c.projectsByName[name]
}
