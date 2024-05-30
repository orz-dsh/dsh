package dsh_core

// region import

type projectImport struct {
	context *appContext
	Entity  *projectImportEntity
	Target  *projectLinkTarget
	project *appProject
}

type projectImportScope string

const (
	projectImportScopeScript projectImportScope = "script"
	projectImportScopeConfig projectImportScope = "config"
)

func newProjectImport(context *appContext, entity *projectImportEntity, target *projectLinkTarget) *projectImport {
	return &projectImport{
		context: context,
		Entity:  entity,
		Target:  target,
	}
}

func (i *projectImport) loadProject() error {
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

// region container

type projectImportContainer struct {
	context       *appContext
	scope         projectImportScope
	ProjectName   string
	ProjectPath   string
	Imports       []*projectImport
	importsByPath map[string]*projectImport
	importsLoaded bool
}

func makeProjectImportContainer(context *appContext, entity *projectEntity, option *projectOption, scope projectImportScope) (container *projectImportContainer, err error) {
	var imports []*projectImportEntity
	if scope == projectImportScopeScript {
		imports = entity.ScriptImports
	} else if scope == projectImportScopeConfig {
		imports = entity.ConfigImports
	} else {
		impossible()
	}
	container = &projectImportContainer{
		context:       context,
		scope:         scope,
		ProjectName:   entity.Name,
		ProjectPath:   entity.Path,
		importsByPath: map[string]*projectImport{},
	}
	for i := 0; i < len(imports); i++ {
		entity := imports[i]
		matched, err := option.evaluator.EvalBoolExpr(entity.match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if err = container.addImport(entity); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *projectImportContainer) addImport(entity *projectImportEntity) (err error) {
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
	imp := newProjectImport(c.context, entity, target)
	if _, exist := c.importsByPath[target.Path]; !exist {
		c.Imports = append(c.Imports, imp)
		c.importsByPath[target.Path] = imp
	}
	return nil
}

func (c *projectImportContainer) loadImports() (err error) {
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
