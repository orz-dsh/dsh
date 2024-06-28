package core

// region projectDependencyEntity

type projectDependencyEntity struct {
	context       *appContext
	ProjectPath   string
	Items         []*projectDependencyItemEntity
	itemPathsDict map[string]bool
	loaded        bool
}

func newProjectDependencyEntity(context *appContext, setting *projectSetting, option *projectOptionEntity) (*projectDependencyEntity, error) {
	import_ := &projectDependencyEntity{
		context:       context,
		ProjectPath:   setting.Dir,
		itemPathsDict: map[string]bool{},
	}
	for i := 0; i < len(setting.Dependency.Items); i++ {
		importSetting := setting.Dependency.Items[i]
		matched, err := option.evaluator.EvalBoolExpr(importSetting.match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if err = import_.addImport(importSetting); err != nil {
			return nil, err
		}
	}
	return import_, nil
}

func (e *projectDependencyEntity) addImport(setting *projectDependencyItemSetting) (err error) {
	target, err := e.context.profile.getProjectLinkTarget(setting.link)
	if err != nil {
		return errW(err, "add import error",
			reason("resolve project link error"),
			kv("setting", setting),
		)
	}
	if target.Path != e.ProjectPath && !e.itemPathsDict[target.Path] {
		item := newProjectImportItemEntity(e.context, target)
		e.Items = append(e.Items, item)
		e.itemPathsDict[target.Path] = true
	}
	return nil
}

func (e *projectDependencyEntity) load() (err error) {
	if e.loaded {
		return nil
	}
	for i := 0; i < len(e.Items); i++ {
		if err = e.Items[i].load(); err != nil {
			return errW(err, "load imports error",
				reason("load import target error"),
			)
		}
	}
	e.loaded = true
	return nil
}

func (e *projectDependencyEntity) inspect() *ProjectDependencyEntityInspection {
	var items []*ProjectDependencyItemEntityInspection
	for i := 0; i < len(e.Items); i++ {
		items = append(items, e.Items[i].inspect())
	}
	return newProjectDependencyEntityInspection(items)
}

// endregion

// region projectDependencyItemEntity

type projectDependencyItemEntity struct {
	context *appContext
	Target  *projectLinkTarget
	project *projectEntity
}

func newProjectImportItemEntity(context *appContext, target *projectLinkTarget) *projectDependencyItemEntity {
	return &projectDependencyItemEntity{
		context: context,
		Target:  target,
	}
}

func (e *projectDependencyItemEntity) load() error {
	if e.project == nil {
		if project, err := e.context.loadProjectByTarget(e.Target); err != nil {
			return err
		} else {
			e.project = project
		}
	}
	return nil
}

func (e *projectDependencyItemEntity) inspect() *ProjectDependencyItemEntityInspection {
	var gitUrl, gitRef string
	if e.Target.Git != nil {
		gitUrl = e.Target.Git.Url
		gitRef = e.Target.Git.Ref
	}
	return newProjectDependencyItemEntityInspection(e.Target.Link.Normalized, e.Target.Path, gitUrl, gitRef)
}

// endregion
