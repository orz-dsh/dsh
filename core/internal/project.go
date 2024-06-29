package internal

import (
	"fmt"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"path/filepath"
	"slices"
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
	evaluator = evaluator.SetData("options", e.option.Items)
	targetNames, err := e.source.makeTargetFiles(evaluator, outputPath, useHardLink)
	if err != nil {
		return nil, ErrW(err, "make scripts error",
			Reason("make sources error"),
			KV("project", e),
		)
	}
	return targetNames, nil
}

func (e *Project) inspect() *ProjectInspection {
	return NewProjectInspection(e.Name, e.Dir, e.option.Inspect(), e.import_.Inspect(), e.source.inspect())
}

// endregion

// region Projects

type Projects struct {
	context            *ApplicationCore
	mainSetting        *ProjectSetting
	extraSettings      []*ProjectSetting
	mainProject        *Project
	extensionProjects  []*Project
	dependencyProjects []*Project
	projects           []*Project
}

func NewProjects(context *ApplicationCore, mainSetting *ProjectSetting, extraSettings []*ProjectSetting) *Projects {
	return &Projects{
		context:       context,
		mainSetting:   mainSetting,
		extraSettings: extraSettings,
	}
}

func (c *Projects) loadImportProjects(project *Project, projectsDict map[string]bool) (projects []*Project, err error) {
	if err = project.loadImports(); err != nil {
		return nil, err
	}

	imp := project.import_
	for i := 0; i < len(imp.Items); i++ {
		p := imp.Items[i].project
		if !projectsDict[p.Dir] {
			projects = append(projects, p)
			projectsDict[p.Dir] = true
		}
	}

	for i := 0; i < len(projects); i++ {
		p1 := projects[i]
		if err = p1.loadImports(); err != nil {
			return nil, err
		}
		imp1 := p1.import_
		for j := 0; j < len(imp1.Items); j++ {
			p2 := imp1.Items[j].project
			if !projectsDict[p2.Dir] {
				projects = append(projects, p2)
				projectsDict[p2.Dir] = true
			}
		}
	}

	return projects, nil
}

func (c *Projects) loadProjects() (err error) {
	if c.mainProject != nil {
		return nil
	}

	// load main project
	var mainProject *Project
	if mainProject, err = c.context.loadProject(c.mainSetting); err != nil {
		return err
	}
	var importProjects []*Project

	// load main project import projects
	projects := []*Project{mainProject}
	projectsDict := map[string]bool{mainProject.Dir: true}
	if _projects, err := c.loadImportProjects(mainProject, projectsDict); err != nil {
		return err
	} else {
		projects = append(projects, _projects...)
		importProjects = append(importProjects, _projects...)
	}

	// load extra projects
	var extraProjects []*Project
	for i := 0; i < len(c.extraSettings); i++ {
		existProject := c.context.getProject(c.extraSettings[i].Name)
		var option *ProjectOption
		if existProject != nil {
			option = existProject.option
		}
		extraProject, err := NewProject(c.context, c.extraSettings[i], option)
		if err != nil {
			return err
		}
		extraProjects = append(extraProjects, extraProject)
		projects = append(projects, extraProject)

		// load extra project import projects
		if _projects, err := c.loadImportProjects(extraProject, projectsDict); err != nil {
			return err
		} else {
			projects = append(projects, _projects...)
			importProjects = append(importProjects, _projects...)
		}
	}

	c.mainProject = mainProject
	c.extensionProjects = extraProjects
	c.dependencyProjects = importProjects
	c.projects = projects
	return nil
}

func (c *Projects) makeConfigs() (configs map[string]any, configsTraces map[string]any, err error) {
	if err = c.loadProjects(); err != nil {
		return nil, nil, ErrW(err, "make configs error",
			Reason("load projects error"),
			// TODO: error
		)
	}

	var contents []*ProjectResourceConfigItemContent
	for i := 0; i < len(c.projects); i++ {
		iContents, err := c.projects[i].loadConfigContents()
		if err != nil {
			return nil, nil, ErrW(err, "make configs error",
				Reason("load config contents error"),
				// TODO: error
				KV("project", c.projects[i]),
			)
		}
		contents = append(contents, iContents...)
	}

	slices.SortStableFunc(contents, func(l, r *ProjectResourceConfigItemContent) int {
		n := l.Order - r.Order
		if n < 0 {
			return 1
		} else if n > 0 {
			return -1
		} else {
			return 0
		}
	})

	configs = map[string]any{}
	configsTraces = map[string]any{}
	for i := 0; i < len(contents); i++ {
		content := contents[i]
		if err = content.merge(configs, configsTraces); err != nil {
			return nil, nil, ErrW(err, "make configs error",
				Reason("merge configs error"),
				KV("file", content.file),
			)
		}
	}
	return configs, configsTraces, nil
}

func (c *Projects) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool, inspectionPath string) ([]string, error) {
	if err := c.loadProjects(); err != nil {
		return nil, ErrW(err, "make scripts error",
			Reason("load projects error"),
			// TODO: error
		)
	}

	if inspectionPath != "" {
		mainProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.main.%s.yml", c.mainProject.Name))
		if err := WriteYamlFile(mainProjectInspectionPath, c.mainProject.inspect()); err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("write project inspection error"),
				KV("project", c.mainProject),
			)
		}
		for i := 0; i < len(c.extensionProjects); i++ {
			extraProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.ext-%d.%s.yml", i+1, c.extensionProjects[i].Name))
			if err := WriteYamlFile(extraProjectInspectionPath, c.extensionProjects[i].inspect()); err != nil {
				return nil, ErrW(err, "make scripts error",
					Reason("write project inspection error"),
					KV("project", c.extensionProjects[i]),
				)
			}
		}
		for i := 0; i < len(c.dependencyProjects); i++ {
			importProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.dep-%d.%s.yml", i+1, c.dependencyProjects[i].Name))
			if err := WriteYamlFile(importProjectInspectionPath, c.dependencyProjects[i].inspect()); err != nil {
				return nil, ErrW(err, "make scripts error",
					Reason("write project inspection error"),
					KV("project", c.dependencyProjects[i]),
				)
			}
		}
	}

	var targetNames []string
	var targetDict = map[string]bool{}
	for i := 0; i < len(c.projects); i++ {
		names, err := c.projects[i].makeScripts(evaluator, outputPath, useHardLink)
		if err != nil {
			return nil, err
		}
		for j := 0; j < len(names); j++ {
			if !targetDict[names[j]] {
				targetNames = append(targetNames, names[j])
				targetDict[names[j]] = true
			}
		}
	}

	return targetNames, nil
}

// endregion
