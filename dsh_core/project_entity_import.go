package dsh_core

// region import

type projectEntityImport struct {
	context *appContext
	Entity  *projectSchemaImport
	Target  *projectLinkTarget
	project *appProject
}

type projectImportScope string

const (
	projectImportScopeScript projectImportScope = "script"
	projectImportScopeConfig projectImportScope = "config"
)

func newProjectImport(context *appContext, entity *projectSchemaImport, target *projectLinkTarget) *projectEntityImport {
	return &projectEntityImport{
		context: context,
		Entity:  entity,
		Target:  target,
	}
}

func (i *projectEntityImport) loadProject() error {
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
	Imports       []*projectEntityImport
	importsByPath map[string]*projectEntityImport
	importsLoaded bool
}

func makeProjectImportContainer(context *appContext, entity *projectSchema, option *projectOption, scope projectImportScope) (container *projectImportContainer, err error) {
	var imports []*projectSchemaImport
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
		importsByPath: map[string]*projectEntityImport{},
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

func (c *projectImportContainer) addImport(entity *projectSchemaImport) (err error) {
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
