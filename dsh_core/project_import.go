package dsh_core

// region import

type projectImport struct {
	context  *appContext
	manifest *projectManifest
	Entity   *projectImportEntity
	Link     *projectResolvedLink
	target   *project
}

type projectImportScope string

const (
	projectImportScopeScript projectImportScope = "script"
	projectImportScopeConfig projectImportScope = "config"
)

func newProjectImport(context *appContext, manifest *projectManifest, entity *projectImportEntity, link *projectResolvedLink) *projectImport {
	return &projectImport{
		context:  context,
		manifest: manifest,
		Entity:   entity,
		Link:     link,
	}
}

func (i *projectImport) loadTarget() error {
	if i.target == nil {
		m, err := i.context.loadProjectManifest(i.Link)
		if err != nil {
			return errW(err, "load import target error",
				kv("reason", "load project manifest error"),
				kv("link", i.Link),
			)
		}
		p, err := i.context.loadProject(m)
		if err != nil {
			return errW(err, "load import target error",
				kv("reason", "load project error"),
				kv("link", i.Link),
			)
		}
		i.target = p
	}
	return nil
}

// endregion

// region container

type projectImportContainer struct {
	context       *appContext
	manifest      *projectManifest
	scope         projectImportScope
	Imports       []*projectImport
	importsByPath map[string]*projectImport
	importsLoaded bool
}

func makeProjectImportContainer(context *appContext, manifest *projectManifest, option *projectOption, scope projectImportScope) (container *projectImportContainer, err error) {
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
		manifest:      manifest,
		scope:         scope,
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
	if resolved.Path == c.manifest.projectPath {
		return nil
	}
	imp := newProjectImport(c.context, c.manifest, entity, resolved)
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
		if err = c.Imports[i].loadTarget(); err != nil {
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
