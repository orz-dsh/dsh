package internal

import (
	"github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
)

// region ProjectDependency

type ProjectDependency struct {
	context       *ApplicationCore
	ProjectPath   string
	Items         []*ProjectDependencyItem
	itemPathsDict map[string]bool
	loaded        bool
}

func NewProjectDependency(context *ApplicationCore, setting *ProjectSetting, option *ProjectOption) (*ProjectDependency, error) {
	import_ := &ProjectDependency{
		context:       context,
		ProjectPath:   setting.Dir,
		itemPathsDict: map[string]bool{},
	}
	for i := 0; i < len(setting.Dependency.Items); i++ {
		importSetting := setting.Dependency.Items[i]
		matched, err := option.evaluator.EvalBoolExpr(importSetting.Match)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		if err = import_.addItem(importSetting); err != nil {
			return nil, err
		}
	}
	return import_, nil
}

func (e *ProjectDependency) addItem(setting *ProjectDependencyItemSetting) (err error) {
	target, err := e.context.Setting.GetProjectLinkTarget(setting.LinkObj)
	if err != nil {
		return ErrW(err, "add import error",
			Reason("resolve project link error"),
			KV("setting", setting),
		)
	}
	if target.Dir != e.ProjectPath && !e.itemPathsDict[target.Dir] {
		item := NewProjectDependencyItem(e.context, target)
		e.Items = append(e.Items, item)
		e.itemPathsDict[target.Dir] = true
	}
	return nil
}

func (e *ProjectDependency) load() (err error) {
	if e.loaded {
		return nil
	}
	for i := 0; i < len(e.Items); i++ {
		if err = e.Items[i].load(); err != nil {
			return ErrW(err, "load imports error",
				Reason("load import target error"),
			)
		}
	}
	e.loaded = true
	return nil
}

func (e *ProjectDependency) Inspect() *ProjectDependencyInspection {
	var items []*ProjectDependencyItemInspection
	for i := 0; i < len(e.Items); i++ {
		items = append(items, e.Items[i].Inspect())
	}
	return NewProjectDependencyInspection(items)
}

// endregion

// region ProjectDependencyItem

type ProjectDependencyItem struct {
	context *ApplicationCore
	Target  *common.ProjectLinkTarget
	project *Project
}

func NewProjectDependencyItem(context *ApplicationCore, target *common.ProjectLinkTarget) *ProjectDependencyItem {
	return &ProjectDependencyItem{
		context: context,
		Target:  target,
	}
}

func (e *ProjectDependencyItem) load() error {
	if e.project == nil {
		if project, err := e.context.loadProjectByTarget(e.Target); err != nil {
			return err
		} else {
			e.project = project
		}
	}
	return nil
}

func (e *ProjectDependencyItem) Inspect() *ProjectDependencyItemInspection {
	var gitUrl, gitRef string
	if e.Target.Git != nil {
		gitUrl = e.Target.Git.Url
		gitRef = e.Target.Git.Ref
	}
	return NewProjectDependencyItemInspection(e.Target.Link.Normalized, e.Target.Dir, gitUrl, gitRef)
}

// endregion
