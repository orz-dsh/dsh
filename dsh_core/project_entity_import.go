package dsh_core

// region projectImportEntity

type projectImportEntity struct {
	context       *appContext
	ProjectName   string
	ProjectPath   string
	Items         []*projectImportItemEntity
	itemPathsDict map[string]bool
	loaded        bool
}

func newProjectImportEntity(context *appContext, setting *projectSetting, option *projectOptionEntity) (*projectImportEntity, error) {
	import_ := &projectImportEntity{
		context:       context,
		ProjectName:   setting.Name,
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

func (e *projectImportEntity) addImport(setting *projectDependencyItemSetting) (err error) {
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

func (e *projectImportEntity) load() (err error) {
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

func (e *projectImportEntity) inspect() *ProjectImportEntityInspection {
	var items []*ProjectImportItemEntityInspection
	for i := 0; i < len(e.Items); i++ {
		items = append(items, e.Items[i].inspect())
	}
	return newProjectImportEntityInspection(items)
}

// endregion

// region projectImportItemEntity

type projectImportItemEntity struct {
	context *appContext
	Target  *projectLinkTarget
	project *projectEntity
}

func newProjectImportItemEntity(context *appContext, target *projectLinkTarget) *projectImportItemEntity {
	return &projectImportItemEntity{
		context: context,
		Target:  target,
	}
}

func (e *projectImportItemEntity) load() error {
	if e.project == nil {
		if project, err := e.context.loadProjectByTarget(e.Target); err != nil {
			return err
		} else {
			e.project = project
		}
	}
	return nil
}

func (e *projectImportItemEntity) inspect() *ProjectImportItemEntityInspection {
	var gitUrl, gitRef string
	if e.Target.Git != nil {
		gitUrl = e.Target.Git.Url
		gitRef = e.Target.Git.Ref
	}
	return newProjectImportItemEntityInspection(e.Target.Link.Normalized, e.Target.Path, gitUrl, gitRef)
}

// endregion

// region ProjectImportEntityInspection

type ProjectImportEntityInspection struct {
	Items []*ProjectImportItemEntityInspection `yaml:"items" toml:"items" json:"items"`
}

func newProjectImportEntityInspection(items []*ProjectImportItemEntityInspection) *ProjectImportEntityInspection {
	return &ProjectImportEntityInspection{
		Items: items,
	}
}

// endregion

// region ProjectImportItemEntityInspection

type ProjectImportItemEntityInspection struct {
	Link   string `yaml:"link" toml:"link" json:"link"`
	Path   string `yaml:"path" toml:"path" json:"path"`
	GitUrl string `yaml:"gitUrl,omitempty" toml:"gitUrl,omitempty" json:"gitUrl,omitempty"`
	GitRef string `yaml:"gitRef,omitempty" toml:"gitRef,omitempty" json:"gitRef,omitempty"`
}

func newProjectImportItemEntityInspection(link string, path string, gitUrl string, gitRef string) *ProjectImportItemEntityInspection {
	return &ProjectImportItemEntityInspection{
		Link:   link,
		Path:   path,
		GitUrl: gitUrl,
		GitRef: gitRef,
	}
}

// endregion
