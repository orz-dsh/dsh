package dsh_core

// region import

type projectImport struct {
	context    *appContext
	manifest   *projectManifest
	Definition *projectImportDefinition
	Link       *projectResolvedLink
	target     *project
}

type projectImportScope string

const (
	projectImportScopeScript projectImportScope = "script"
	projectImportScopeConfig projectImportScope = "config"
)

func newProjectImport(context *appContext, manifest *projectManifest, definition *projectImportDefinition, link *projectResolvedLink) *projectImport {
	return &projectImport{
		context:    context,
		manifest:   manifest,
		Definition: definition,
		Link:       link,
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

func makeProjectImportContainer(context *appContext, manifest *projectManifest, scope projectImportScope) (container *projectImportContainer, err error) {
	var definitions []*projectImportDefinition
	if scope == projectImportScopeScript {
		definitions = manifest.Script.importDefinitions
		if context.isMainProject(manifest) {
			definitions = append(definitions, context.Profile.getProjectScriptImportDefinitions()...)
		}
	} else if scope == projectImportScopeConfig {
		definitions = manifest.Config.importDefinitions
		if context.isMainProject(manifest) {
			definitions = append(definitions, context.Profile.getProjectConfigImportDefinitions()...)
		}
	} else {
		impossible()
	}
	container = &projectImportContainer{
		context:       context,
		manifest:      manifest,
		scope:         scope,
		importsByPath: make(map[string]*projectImport),
	}
	for i := 0; i < len(definitions); i++ {
		if err = container.addImport(definitions[i]); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *projectImportContainer) addImport(definition *projectImportDefinition) (err error) {
	if definition.match != nil {
		matched, err := c.context.Option.evalMatch(c.manifest, definition.match)
		if err != nil {
			return errW(err, "add import error",
				reason("eval match error"),
				kv("scope", c.scope),
				kv("definition", definition),
			)
		}
		if !matched {
			return nil
		}
	}
	resolved, err := c.context.Profile.resolveProjectLink(definition.Link)
	if err != nil {
		return errW(err, "add import error",
			reason("resolve project link error"),
			kv("scope", c.scope),
			kv("definition", definition),
		)
	}
	if resolved.Path == c.manifest.projectPath {
		return nil
	}
	imp := newProjectImport(c.context, c.manifest, definition, resolved)
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
