package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"slices"
)

// region projectInstance

type projectInstance struct {
	Name       string
	Path       string
	context    *appContext
	option     *projectOptionInstance
	script     *projectScriptInstance
	config     *projectConfigInstance
	extra      bool
	extraIndex int
}

func newProjectInstance(context *appContext, setting *projectSetting) (instance *projectInstance, err error) {
	context.logger.InfoDesc("load project", kv("name", setting.Name))
	option, err := makeProjectOption(context, setting)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("make project option error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Path),
		)
	}
	script, err := newProjectScriptInstance(context, setting, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project script error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Path),
		)
	}
	config, err := newProjectConfigInstance(context, setting, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project config error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Path),
		)
	}
	instance = &projectInstance{
		Name:    setting.Name,
		Path:    setting.Path,
		context: context,
		option:  option,
		script:  script,
		config:  config,
	}
	return instance, nil
}

func newProjectInstanceFromExtraSetting(context *appContext, setting *projectSetting, option *projectOptionInstance, extraIndex int) (instance *projectInstance, err error) {
	if option == nil {
		option, err = makeProjectOption(context, setting)
		if err != nil {
			return nil, errW(err, "load project error",
				reason("make project option error"),
				kv("projectName", setting.Name),
				kv("projectPath", setting.Path),
			)
		}
	}
	script, err := newProjectScriptInstance(context, setting, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project script error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Path),
		)
	}
	config, err := newProjectConfigInstance(context, setting, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project config error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Path),
		)
	}
	instance = &projectInstance{
		Name:       setting.Name,
		Path:       setting.Path,
		context:    context,
		option:     option,
		script:     script,
		config:     config,
		extra:      true,
		extraIndex: extraIndex,
	}
	return instance, nil
}

func (i *projectInstance) getImportContainer(scope projectImportScope) *projectImportInstanceContainer {
	if scope == projectImportScopeScript {
		return i.script.ImportContainer
	} else if scope == projectImportScopeConfig {
		return i.config.ImportContainer
	} else {
		impossible()
	}
	return nil
}

func (i *projectInstance) loadImports(scope projectImportScope) error {
	return i.getImportContainer(scope).loadImports()
}

func (i *projectInstance) loadConfigContents() ([]*projectConfigContentInstance, error) {
	return i.config.SourceContainer.loadContents()
}

func (i *projectInstance) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool, inspectionPath string) ([]string, error) {
	if inspectionPath != "" {
		projectInspectionPath := ""
		if i.extra {
			projectInspectionPath = filepath.Join(inspectionPath, fmt.Sprintf("extra-project-%d.%s.yml", i.extraIndex, i.Name))
		} else {
			projectInspectionPath = filepath.Join(inspectionPath, fmt.Sprintf("project.%s.yml", i.Name))
		}
		projectInspection := i.inspect()
		if err := dsh_utils.WriteYamlFile(projectInspectionPath, projectInspection); err != nil {
			return nil, errW(err, "make scripts error",
				reason("write project inspection error"),
				kv("project", i),
			)
		}
	}

	evaluator = evaluator.SetData("options", i.option.Items)
	targetNames, err := i.script.SourceContainer.makeSources(evaluator, outputPath, useHardLink)
	if err != nil {
		return nil, errW(err, "make scripts error",
			reason("make sources error"),
			kv("project", i),
		)
	}
	return targetNames, nil
}

func (i *projectInstance) inspect() *ProjectInstanceInspection {
	return newProjectInstanceInspection(i.Name, i.Path, i.option.inspect(), i.script.inspect(), i.config.inspect())
}

// endregion

// region projectInstanceContainer

type projectInstanceContainer struct {
	context        *appContext
	mainSetting    *projectSetting
	extraSettings  []*projectSetting
	mainProject    *projectInstance
	extraProjects  []*projectInstance
	scriptProjects []*projectInstance
	configProjects []*projectInstance
}

func newProjectInstanceContainerTest(context *appContext, mainSetting *projectSetting, extraSettings []*projectSetting) *projectInstanceContainer {
	return &projectInstanceContainer{
		context:       context,
		mainSetting:   mainSetting,
		extraSettings: extraSettings,
	}
}

func (c *projectInstanceContainer) loadImportProjects(scope projectImportScope, project *projectInstance, projectsDict map[string]bool) (projects []*projectInstance, err error) {
	if err = project.loadImports(scope); err != nil {
		return nil, err
	}

	pic := project.getImportContainer(scope)
	for i := 0; i < len(pic.Imports); i++ {
		p := pic.Imports[i].project
		if !projectsDict[p.Path] {
			projects = append(projects, p)
			projectsDict[p.Path] = true
		}
	}

	for i := 0; i < len(projects); i++ {
		p1 := projects[i]
		if err = p1.loadImports(scope); err != nil {
			return nil, err
		}
		pic1 := p1.getImportContainer(scope)
		for j := 0; j < len(pic1.Imports); j++ {
			p2 := pic1.Imports[j].project
			if !projectsDict[p2.Path] {
				projects = append(projects, p2)
				projectsDict[p2.Path] = true
			}
		}
	}

	return projects, nil
}

func (c *projectInstanceContainer) loadProjects() (err error) {
	if c.mainProject != nil {
		return nil
	}

	// load main project
	var mainProject *projectInstance
	if mainProject, err = c.context.loadProject(c.mainSetting); err != nil {
		return err
	}

	// load main project script import projects
	scriptProjects := []*projectInstance{mainProject}
	scriptProjectsDict := map[string]bool{mainProject.Path: true}
	if projects, err := c.loadImportProjects(projectImportScopeScript, mainProject, scriptProjectsDict); err != nil {
		return err
	} else {
		scriptProjects = append(scriptProjects, projects...)
	}

	// load main project config import projects
	configProjects := []*projectInstance{mainProject}
	configProjectsDict := map[string]bool{mainProject.Path: true}
	if projects, err := c.loadImportProjects(projectImportScopeConfig, mainProject, configProjectsDict); err != nil {
		return err
	} else {
		configProjects = append(configProjects, projects...)
	}

	// load extra projects
	var extraProjects []*projectInstance
	for i := 0; i < len(c.extraSettings); i++ {
		existProject := c.context.getProject(c.extraSettings[i].Name)
		var option *projectOptionInstance
		if existProject != nil {
			option = existProject.option
		}
		extraProject, err := newProjectInstanceFromExtraSetting(c.context, c.extraSettings[i], option, i)
		if err != nil {
			return err
		}
		extraProjects = append(extraProjects, extraProject)
		scriptProjects = append(scriptProjects, extraProject)
		configProjects = append(configProjects, extraProject)

		// load extra project script import projects
		if projects, err := c.loadImportProjects(projectImportScopeScript, extraProject, scriptProjectsDict); err != nil {
			return err
		} else {
			scriptProjects = append(scriptProjects, projects...)
		}

		// load extra project config import projects
		if projects, err := c.loadImportProjects(projectImportScopeConfig, extraProject, configProjectsDict); err != nil {
			return err
		} else {
			configProjects = append(configProjects, projects...)
		}
	}

	c.mainProject = mainProject
	c.extraProjects = extraProjects
	c.scriptProjects = scriptProjects
	c.configProjects = configProjects
	return nil
}

func (c *projectInstanceContainer) makeConfigs() (configs map[string]any, configsTraces map[string]any, err error) {
	if err = c.loadProjects(); err != nil {
		return nil, nil, errW(err, "make configs error",
			reason("load projects error"),
			// TODO: error
		)
	}

	var contents []*projectConfigContentInstance
	for i := 0; i < len(c.configProjects); i++ {
		iContents, err := c.configProjects[i].loadConfigContents()
		if err != nil {
			return nil, nil, errW(err, "make configs error",
				reason("load config contents error"),
				// TODO: error
				kv("project", c.configProjects[i]),
			)
		}
		contents = append(contents, iContents...)
	}

	slices.SortStableFunc(contents, func(l, r *projectConfigContentInstance) int {
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
				kv("sourcePath", content.sourcePath),
			)
		}
	}
	return configs, configsTraces, nil
}

func (c *projectInstanceContainer) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool, inspectionPath string) ([]string, error) {
	if err := c.loadProjects(); err != nil {
		return nil, errW(err, "make scripts error",
			reason("load projects error"),
			// TODO: error
		)
	}

	var targetNames []string
	var targetDict = map[string]bool{}
	for i := 0; i < len(c.scriptProjects); i++ {
		names, err := c.scriptProjects[i].makeScripts(evaluator, outputPath, useHardLink, inspectionPath)
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

// region ProjectInstanceInspection

type ProjectInstanceInspection struct {
	Name   string                           `yaml:"name" toml:"name" json:"name"`
	Path   string                           `yaml:"path" toml:"path" json:"path"`
	Option *ProjectOptionInstanceInspection `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Script *ProjectScriptInstanceInspection `yaml:"script,omitempty" toml:"script,omitempty" json:"script,omitempty"`
	Config *ProjectConfigInstanceInspection `yaml:"config,omitempty" toml:"config,omitempty" json:"config,omitempty"`
}

func newProjectInstanceInspection(name string, path string, option *ProjectOptionInstanceInspection, script *ProjectScriptInstanceInspection, config *ProjectConfigInstanceInspection) *ProjectInstanceInspection {
	return &ProjectInstanceInspection{
		Name:   name,
		Path:   path,
		Option: option,
		Script: script,
		Config: config,
	}
}

// endregion
