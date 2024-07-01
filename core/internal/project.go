package internal

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
)

// region Project

type Project struct {
	Name    string
	Dir     string
	context *ApplicationCore
	option  *ProjectOption
	import_ *ProjectDependency
	source  *ProjectResource
}

func NewProject(context *ApplicationCore, setting *ProjectSetting, option *ProjectOption) (_ *Project, err error) {
	context.Logger.InfoDesc("load project", KV("name", setting.Name))
	if option == nil {
		option, err = NewProjectOption(context, setting)
		if err != nil {
			return nil, ErrW(err, "load project error",
				Reason("new project option error"),
				KV("projectName", setting.Name),
				KV("projectPath", setting.Dir),
			)
		}
	}
	import_, err := NewProjectDependency(context, setting, option)
	if err != nil {
		return nil, ErrW(err, "load project error",
			Reason("new project import error"),
			KV("projectName", setting.Name),
			KV("projectPath", setting.Dir),
		)
	}
	source, err := NewProjectResource(context, setting, option)
	if err != nil {
		return nil, ErrW(err, "load project error",
			Reason("new project source error"),
			KV("projectName", setting.Name),
			KV("projectPath", setting.Dir),
		)
	}
	project := &Project{
		Name:    setting.Name,
		Dir:     setting.Dir,
		context: context,
		option:  option,
		import_: import_,
		source:  source,
	}
	return project, nil
}

func (e *Project) loadImports() error {
	return e.import_.load()
}

func (e *Project) loadConfigContents() ([]*ProjectResourceConfigItemContent, error) {
	return e.source.loadConfigFiles()
}

func (e *Project) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool) ([]string, error) {
	evaluator = evaluator.SetData("option", e.option.Items)
	targetNames, err := e.source.makeTargetFiles(evaluator, outputPath, useHardLink)
	if err != nil {
		return nil, ErrW(err, "make scripts error",
			Reason("make sources error"),
			KV("project", e),
		)
	}
	return targetNames, nil
}

func (e *Project) Inspect() *ProjectInspection {
	return NewProjectInspection(e.Name, e.Dir, e.option.Inspect(), e.import_.Inspect(), e.source.inspect())
}

// endregion
