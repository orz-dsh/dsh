package dsh_core

// region import

type projectImport struct {
	context *appContext
	Entity  *projectImportEntity
	Link    *projectResolvedLink
	project *Project
}

type projectImportScope string

const (
	projectImportScopeScript projectImportScope = "script"
	projectImportScopeConfig projectImportScope = "config"
)

func newProjectImport(context *appContext, entity *projectImportEntity, link *projectResolvedLink) *projectImport {
	return &projectImport{
		context: context,
		Entity:  entity,
		Link:    link,
	}
}

func (i *projectImport) loadProject() error {
	if i.project == nil {
		m, err := i.context.loadProjectManifest(i.Link)
		if err != nil {
			return errW(err, "load import target error",
				kv("reason", "load project manifest error"),
				kv("link", i.Link),
			)
		}
		project, err := i.context.loadProject(m)
		if err != nil {
			return errW(err, "load import target error",
				kv("reason", "load project error"),
				kv("link", i.Link),
			)
		}
		i.project = project
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

func makeProjectImportContainer(context *appContext, manifest *ProjectManifest, option *projectOption, scope projectImportScope) (container *projectImportContainer, err error) {
	var entities []*projectImportEntity
	if scope == projectImportScopeScript {
		entities = manifest.Script.importEntities
		if context.isMainProject(manifest) {
			entities = append(entities, context.profile.projectScriptImportEntities...)
		}
	} else if scope == projectImportScopeConfig {
		entities = manifest.Config.importEntities
		if context.isMainProject(manifest) {
			entities = append(entities, context.profile.projectConfigImportEntities...)
		}
	} else {
		impossible()
	}
	container = &projectImportContainer{
		context:       context,
		scope:         scope,
		ProjectName:   manifest.projectName,
		ProjectPath:   manifest.projectPath,
		importsByPath: map[string]*projectImport{},
	}
	for i := 0; i < len(entities); i++ {
		entity := entities[i]
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
	resolved, err := c.context.profile.resolveProjectLink(entity.link)
	if err != nil {
		return errW(err, "add import error",
			reason("resolve project link error"),
			kv("scope", c.scope),
			kv("entity", entity),
		)
	}
	if resolved.Path == c.ProjectPath {
		return nil
	}
	imp := newProjectImport(c.context, entity, resolved)
	if _, exist := c.importsByPath[resolved.Path]; !exist {
		c.Imports = append(c.Imports, imp)
		c.importsByPath[resolved.Path] = imp
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
