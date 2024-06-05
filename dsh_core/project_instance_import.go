package dsh_core

// region projectImportInstance

type projectImportInstance struct {
	context *appContext
	Setting *projectImportSetting
	Target  *projectLinkTarget
	project *projectInstance
}

type projectImportScope string

const (
	projectImportScopeScript projectImportScope = "script"
	projectImportScopeConfig projectImportScope = "config"
)

func newProjectImportInstance(context *appContext, setting *projectImportSetting, target *projectLinkTarget) *projectImportInstance {
	return &projectImportInstance{
		context: context,
		Setting: setting,
		Target:  target,
	}
}

func (i *projectImportInstance) loadProject() error {
	if i.project == nil {
		if project, err := i.context.loadProjectByTarget(i.Target); err != nil {
			return err
		} else {
			i.project = project
		}
	}
	return nil
}

// endregion

// region projectImportInstanceContainer

type projectImportInstanceContainer struct {
	context       *appContext
	scope         projectImportScope
	ProjectName   string
	ProjectPath   string
	Imports       []*projectImportInstance
	importsByPath map[string]*projectImportInstance
	importsLoaded bool
}

func newProjectImportInstanceContainer(context *appContext, setting *projectSetting, option *projectOptionInstance, scope projectImportScope) (*projectImportInstanceContainer, error) {
	var importSettings []*projectImportSetting
	if scope == projectImportScopeScript {
		importSettings = setting.ScriptImportSettings
	} else if scope == projectImportScopeConfig {
		importSettings = setting.ConfigImportSettings
	} else {
		impossible()
	}
	container := &projectImportInstanceContainer{
		context:       context,
		scope:         scope,
		ProjectName:   setting.Name,
		ProjectPath:   setting.Path,
		importsByPath: map[string]*projectImportInstance{},
	}
	for i := 0; i < len(importSettings); i++ {
		importSetting := importSettings[i]
		matched, err := option.evaluator.EvalBoolExpr(importSetting.match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if err = container.addImport(importSetting); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *projectImportInstanceContainer) addImport(entity *projectImportSetting) (err error) {
	target, err := c.context.profile.getProjectLinkTarget(entity.link)
	if err != nil {
		return errW(err, "add import error",
			reason("resolve project link error"),
			kv("scope", c.scope),
			kv("entity", entity),
		)
	}
	if target.Path == c.ProjectPath {
		return nil
	}
	imp := newProjectImportInstance(c.context, entity, target)
	if _, exist := c.importsByPath[target.Path]; !exist {
		c.Imports = append(c.Imports, imp)
		c.importsByPath[target.Path] = imp
	}
	return nil
}

func (c *projectImportInstanceContainer) loadImports() (err error) {
	if c.importsLoaded {
		return nil
	}
	for i := 0; i < len(c.Imports); i++ {
		if err = c.Imports[i].loadProject(); err != nil {
			return errW(err, "load imports error",
				reason("load import target error"),
				kv("scope", c.scope),
			)
		}
	}
	c.importsLoaded = true
	return nil
}

// endregion
