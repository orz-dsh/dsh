package internal

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"path/filepath"
	"time"
)

// region ApplicationCore

type ApplicationCore struct {
	Logger                  *Logger
	Workspace               *WorkspaceCore
	Evaluator               *Evaluator
	Setting                 *ApplicationSetting
	Option                  *ApplicationOption
	Config                  *ApplicationConfig
	MainProjectSetting      *ProjectSetting
	AdditionProjectSettings []*ProjectSetting
	MainProject             *Project
	AdditionProjects        []*Project
	DependencyProjects      []*Project
	Projects                []*Project
	projectsByName          map[string]*Project
}

func NewApplicationCore(workspace *WorkspaceCore, setting *ApplicationSetting, link string) (*ApplicationCore, error) {
	mainProjectSetting, err := setting.GetProjectEntityByRawLink(link)
	if err != nil {
		return nil, err
	}

	evaluator := workspace.Evaluator.MergeData("local", map[string]any{
		"project_name": mainProjectSetting.Name,
		"project_dir":  mainProjectSetting.Dir,
	})

	arguments, err := setting.Argument.GetArguments(evaluator)
	if err != nil {
		return nil, err
	}

	additionProjectSettings, err := setting.GetAdditionProjectSettings(evaluator)
	if err != nil {
		return nil, err
	}

	option := NewApplicationOption(mainProjectSetting.Name, workspace.SystemInfo, evaluator, arguments)
	core := &ApplicationCore{
		Logger:                  workspace.Logger,
		Workspace:               workspace,
		Evaluator:               evaluator,
		Setting:                 setting,
		Option:                  option,
		MainProjectSetting:      mainProjectSetting,
		AdditionProjectSettings: additionProjectSettings,
		projectsByName:          map[string]*Project{},
	}
	return core, nil
}

func (a *ApplicationCore) loadProject(setting *ProjectSetting) (project *Project, err error) {
	if existProject, exist := a.projectsByName[setting.Name]; exist {
		return existProject, nil
	}
	if project, err = NewProject(a, setting, nil); err != nil {
		return nil, err
	}
	a.projectsByName[setting.Name] = project
	return project, nil
}

func (a *ApplicationCore) loadProjectByTarget(target *ProjectLinkTarget) (project *Project, err error) {
	setting, err := a.Setting.GetProjectSettingByLinkTarget(target)
	if err != nil {
		return nil, ErrW(err, "load project error",
			KV("reason", "load project setting error"),
			KV("target", target),
		)
	}
	project, err = a.loadProject(setting)
	if err != nil {
		return nil, ErrW(err, "load project error",
			KV("target", target),
		)
	}
	return project, nil
}

func (a *ApplicationCore) getProject(name string) *Project {
	return a.projectsByName[name]
}

func (a *ApplicationCore) loadImportProjects(project *Project, projectsDict map[string]bool) (projects []*Project, err error) {
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

func (a *ApplicationCore) loadProjects() (err error) {
	if a.MainProject != nil {
		return nil
	}

	// load main project
	var mainProject *Project
	if mainProject, err = a.loadProject(a.MainProjectSetting); err != nil {
		return err
	}
	var importProjects []*Project

	// load main project import projects
	projects := []*Project{mainProject}
	projectsDict := map[string]bool{mainProject.Dir: true}
	if _projects, err := a.loadImportProjects(mainProject, projectsDict); err != nil {
		return err
	} else {
		projects = append(projects, _projects...)
		importProjects = append(importProjects, _projects...)
	}

	// load extra projects
	var extraProjects []*Project
	for i := 0; i < len(a.AdditionProjectSettings); i++ {
		existProject := a.getProject(a.AdditionProjectSettings[i].Name)
		var option *ProjectOption
		if existProject != nil {
			option = existProject.option
		}
		extraProject, err := NewProject(a, a.AdditionProjectSettings[i], option)
		if err != nil {
			return err
		}
		extraProjects = append(extraProjects, extraProject)
		projects = append(projects, extraProject)

		// load extra project import projects
		if _projects, err := a.loadImportProjects(extraProject, projectsDict); err != nil {
			return err
		} else {
			projects = append(projects, _projects...)
			importProjects = append(importProjects, _projects...)
		}
	}

	a.MainProject = mainProject
	a.AdditionProjects = extraProjects
	a.DependencyProjects = importProjects
	a.Projects = projects
	return nil
}

func (a *ApplicationCore) LoadConfig() error {
	if a.Config != nil {
		return nil
	}

	startTime := time.Now()
	a.Logger.Info("make config start")

	if err := a.loadProjects(); err != nil {
		return ErrW(err, "make config error",
			Reason("load projects error"),
			// TODO: error
		)
	}

	config, err := NewApplicationConfig(a.Evaluator, a.Projects)
	if err != nil {
		return ErrW(err, "make config error",
			Reason("make config error"),
			// TODO: error
		)
	}
	a.Config = config

	a.Logger.InfoDesc("make config finish", KV("elapsed", time.Since(startTime)))
	return nil
}

func (a *ApplicationCore) MakeArtifact(options MakeArtifactOptions) (artifact *ArtifactCore, err error) {
	if err = a.LoadConfig(); err != nil {
		return nil, err
	}

	startTime := time.Now()
	a.Logger.Info("make artifact start")
	outputDir := options.OutputDir
	if outputDir == "" {
		outputDir, err = a.Workspace.MakeOutputDir(a.MainProject.Name)
		if err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("make output path error"),
			)
		}
	} else {
		absPath, err := filepath.Abs(outputDir)
		if err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("get abs-path error"),
				KV("path", outputDir),
			)
		}
		outputDir = absPath
		if options.OutputDirClear {
			if err = ClearDir(outputDir); err != nil {
				return nil, ErrW(err, "make scripts error",
					Reason("clear output dir error"),
					KV("path", outputDir),
				)
			}
		}
	}

	if err = a.loadProjects(); err != nil {
		return nil, ErrW(err, "make scripts error",
			Reason("load projects error"),
			// TODO: error
		)
	}

	if options.InspectSerializer != nil {
		if err = a.SaveInspection(options.InspectSerializer, outputDir); err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("save inspection error"),
			)
		}
	}

	evaluator := a.Config.Evaluator.MergeFuncs(newProjectScriptTemplateFuncs())
	var targetNames []string
	targetNamesDict := map[string]bool{}
	for i := 0; i < len(a.Projects); i++ {
		names, err := a.Projects[i].makeScripts(evaluator, outputDir, options.UseHardLink)
		if err != nil {
			return nil, err
		}
		for j := 0; j < len(names); j++ {
			if !targetNamesDict[names[j]] {
				targetNames = append(targetNames, names[j])
				targetNamesDict[names[j]] = true
			}
		}
	}

	a.Logger.InfoDesc("make artifact finish", KV("elapsed", time.Since(startTime)))
	return NewArtifactCore(a, outputDir, targetNames, targetNamesDict), nil
}

// endregion
