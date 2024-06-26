package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"slices"
)

// region projectEntity

type projectEntity struct {
	Name    string
	Path    string
	context *appContext
	option  *projectOptionEntity
	import_ *projectImportEntity
	source  *projectSourceEntity
}

func newProjectEntity(context *appContext, setting *projectSetting, option *projectOptionEntity) (_ *projectEntity, err error) {
	context.logger.InfoDesc("load project", kv("name", setting.Name))
	if option == nil {
		option, err = newProjectOptionEntity(context, setting)
		if err != nil {
			return nil, errW(err, "load project error",
				reason("new project option error"),
				kv("projectName", setting.Name),
				kv("projectPath", setting.Dir),
			)
		}
	}
	import_, err := newProjectImportEntity(context, setting, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("new project import error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Dir),
		)
	}
	source, err := newProjectSourceEntity(context, setting, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("new project source error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Dir),
		)
	}
	project := &projectEntity{
		Name:    setting.Name,
		Path:    setting.Dir,
		context: context,
		option:  option,
		import_: import_,
		source:  source,
	}
	return project, nil
}

func (e *projectEntity) loadImports() error {
	return e.import_.load()
}

func (e *projectEntity) loadConfigContents() ([]*projectSourceConfigContentEntity, error) {
	return e.source.loadConfigFiles()
}

func (e *projectEntity) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool, inspectionPath string) ([]string, error) {
	evaluator = evaluator.SetData("options", e.option.Items)
	targetNames, err := e.source.makeTargetFiles(evaluator, outputPath, useHardLink)
	if err != nil {
		return nil, errW(err, "make scripts error",
			reason("make sources error"),
			kv("project", e),
		)
	}
	return targetNames, nil
}

func (e *projectEntity) inspect() *ProjectEntityInspection {
	return newProjectEntityInspection(e.Name, e.Path, e.option.inspect(), e.import_.inspect(), e.source.inspect())
}

// endregion

// region projectEntityContainer

type projectEntityContainer struct {
	context        *appContext
	mainSetting    *projectSetting
	extraSettings  []*projectSetting
	mainProject    *projectEntity
	extraProjects  []*projectEntity
	importProjects []*projectEntity
	projects       []*projectEntity
}

func newProjectEntityContainer(context *appContext, mainSetting *projectSetting, extraSettings []*projectSetting) *projectEntityContainer {
	return &projectEntityContainer{
		context:       context,
		mainSetting:   mainSetting,
		extraSettings: extraSettings,
	}
}

func (c *projectEntityContainer) loadImportProjects(project *projectEntity, projectsDict map[string]bool) (projects []*projectEntity, err error) {
	if err = project.loadImports(); err != nil {
		return nil, err
	}

	imp := project.import_
	for i := 0; i < len(imp.Items); i++ {
		p := imp.Items[i].project
		if !projectsDict[p.Path] {
			projects = append(projects, p)
			projectsDict[p.Path] = true
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
			if !projectsDict[p2.Path] {
				projects = append(projects, p2)
				projectsDict[p2.Path] = true
			}
		}
	}

	return projects, nil
}

func (c *projectEntityContainer) loadProjects() (err error) {
	if c.mainProject != nil {
		return nil
	}

	// load main project
	var mainProject *projectEntity
	if mainProject, err = c.context.loadProject(c.mainSetting); err != nil {
		return err
	}
	var importProjects []*projectEntity

	// load main project import projects
	projects := []*projectEntity{mainProject}
	projectsDict := map[string]bool{mainProject.Path: true}
	if _projects, err := c.loadImportProjects(mainProject, projectsDict); err != nil {
		return err
	} else {
		projects = append(projects, _projects...)
		importProjects = append(importProjects, _projects...)
	}

	// load extra projects
	var extraProjects []*projectEntity
	for i := 0; i < len(c.extraSettings); i++ {
		existProject := c.context.getProject(c.extraSettings[i].Name)
		var option *projectOptionEntity
		if existProject != nil {
			option = existProject.option
		}
		extraProject, err := newProjectEntity(c.context, c.extraSettings[i], option)
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
	c.extraProjects = extraProjects
	c.importProjects = importProjects
	c.projects = projects
	return nil
}

func (c *projectEntityContainer) makeConfigs() (configs map[string]any, configsTraces map[string]any, err error) {
	if err = c.loadProjects(); err != nil {
		return nil, nil, errW(err, "make configs error",
			reason("load projects error"),
			// TODO: error
		)
	}

	var contents []*projectSourceConfigContentEntity
	for i := 0; i < len(c.projects); i++ {
		iContents, err := c.projects[i].loadConfigContents()
		if err != nil {
			return nil, nil, errW(err, "make configs error",
				reason("load config contents error"),
				// TODO: error
				kv("project", c.projects[i]),
			)
		}
		contents = append(contents, iContents...)
	}

	slices.SortStableFunc(contents, func(l, r *projectSourceConfigContentEntity) int {
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
			return nil, nil, errW(err, "make configs error",
				reason("merge configs error"),
				kv("path", content.path),
			)
		}
	}
	return configs, configsTraces, nil
}

func (c *projectEntityContainer) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool, inspectionPath string) ([]string, error) {
	if err := c.loadProjects(); err != nil {
		return nil, errW(err, "make scripts error",
			reason("load projects error"),
			// TODO: error
		)
	}

	if inspectionPath != "" {
		mainProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.main.%s.yml", c.mainProject.Name))
		if err := dsh_utils.WriteYamlFile(mainProjectInspectionPath, c.mainProject.inspect()); err != nil {
			return nil, errW(err, "make scripts error",
				reason("write project inspection error"),
				kv("project", c.mainProject),
			)
		}
		for i := 0; i < len(c.extraProjects); i++ {
			extraProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.extra-%d.%s.yml", i+1, c.extraProjects[i].Name))
			if err := dsh_utils.WriteYamlFile(extraProjectInspectionPath, c.extraProjects[i].inspect()); err != nil {
				return nil, errW(err, "make scripts error",
					reason("write project inspection error"),
					kv("project", c.extraProjects[i]),
				)
			}
		}
		for i := 0; i < len(c.importProjects); i++ {
			importProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.import-%d.%s.yml", i+1, c.importProjects[i].Name))
			if err := dsh_utils.WriteYamlFile(importProjectInspectionPath, c.importProjects[i].inspect()); err != nil {
				return nil, errW(err, "make scripts error",
					reason("write project inspection error"),
					kv("project", c.importProjects[i]),
				)
			}
		}
	}

	var targetNames []string
	var targetDict = map[string]bool{}
	for i := 0; i < len(c.projects); i++ {
		names, err := c.projects[i].makeScripts(evaluator, outputPath, useHardLink, inspectionPath)
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

// region ProjectEntityInspection

type ProjectEntityInspection struct {
	Name   string                         `yaml:"name" toml:"name" json:"name"`
	Path   string                         `yaml:"path" toml:"path" json:"path"`
	Option *ProjectOptionEntityInspection `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Import *ProjectImportEntityInspection `yaml:"import,omitempty" toml:"import,omitempty" json:"import,omitempty"`
	Source *ProjectSourceEntityInspection `yaml:"source,omitempty" toml:"source,omitempty" json:"source,omitempty"`
}

func newProjectEntityInspection(name string, path string, option *ProjectOptionEntityInspection, import_ *ProjectImportEntityInspection, source *ProjectSourceEntityInspection) *ProjectEntityInspection {
	return &ProjectEntityInspection{
		Name:   name,
		Path:   path,
		Option: option,
		Import: import_,
		Source: source,
	}
}

// endregion
